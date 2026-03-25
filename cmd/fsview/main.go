package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/iceisfun/gotui/pkg/input"
	"github.com/iceisfun/gotui/pkg/layout"
	"github.com/iceisfun/gotui/pkg/render"
	"github.com/iceisfun/gotui/pkg/term"
	"github.com/iceisfun/gotui/pkg/text"
	"github.com/iceisfun/gotui/pkg/widget"
)

// fsEntry stores filesystem metadata attached to each tree node.
type fsEntry struct {
	Path    string
	Info    fs.FileInfo
	IsDir   bool
	Loaded  bool // directories only: children have been read from disk
	Hidden  bool // name starts with '.'
	Symlink bool
	LinkTgt string
}

// fsViewer holds the application state.
type fsViewer struct {
	rootPath   string
	showHidden bool

	tree      *widget.Tree
	wrapper   *treeWrapper
	treePanel *widget.Panel
	preview   *widget.TextView
	prevPanel *widget.Panel
	statusBar *statusWidget
	app       *layout.App
}

// ---------------------------------------------------------------------------
// statusWidget — a one-line bar at the bottom of the screen.
// ---------------------------------------------------------------------------

type statusWidget struct {
	line text.StyledLine
}

func (s *statusWidget) Render(v *render.View) {
	if v.Height() < 1 {
		return
	}
	bg := text.Style{}.WithBg(text.Blue()).WithFg(text.White()).Bold()
	for x := 0; x < v.Width(); x++ {
		v.SetRune(x, 0, ' ', bg)
	}
	col := 0
	for _, span := range s.line {
		st := span.Style
		st.Bg = text.Blue()
		col += v.WriteString(col, 0, span.Text, st)
	}
}

func (s *statusWidget) SetText(line text.StyledLine) {
	s.line = line
}

// ---------------------------------------------------------------------------
// treeWrapper — sits between Panel and Tree to intercept events for lazy
// loading and to update the preview when the cursor moves.
// ---------------------------------------------------------------------------

type treeWrapper struct {
	fv   *fsViewer
	tree *widget.Tree
}

func (tw *treeWrapper) Render(v *render.View) {
	tw.ensureLoaded()
	tw.tree.Render(v)
}

func (tw *treeWrapper) HandleEvent(ev input.Event) bool {
	consumed := tw.tree.HandleEvent(ev)
	if consumed {
		tw.ensureLoaded()
		// After every consumed key event, update the preview to reflect
		// whatever node the cursor now sits on.
		if node := tw.cursorNode(); node != nil {
			tw.fv.showPreview(node)
		}
	}
	return consumed
}

func (tw *treeWrapper) Focus()                    { tw.tree.Focus() }
func (tw *treeWrapper) Blur()                     { tw.tree.Blur() }
func (tw *treeWrapper) IsFocused() bool           { return tw.tree.IsFocused() }
func (tw *treeWrapper) ContentSize() (int, int)   { return tw.tree.ContentSize() }
func (tw *treeWrapper) ScrollOffset() (int, int)  { return tw.tree.ScrollOffset() }
func (tw *treeWrapper) SetScrollOffset(x, y int)  { tw.tree.SetScrollOffset(x, y) }

// ensureLoaded walks all expanded-but-unloaded directory nodes and loads them.
func (tw *treeWrapper) ensureLoaded() {
	for _, root := range tw.tree.Roots {
		tw.loadExpanded(root)
	}
}

func (tw *treeWrapper) loadExpanded(node *widget.TreeNode) {
	if node == nil {
		return
	}
	entry, ok := node.Data.(*fsEntry)
	if !ok {
		return
	}
	if entry.IsDir && node.Expanded && !entry.Loaded {
		tw.fv.loadChildren(node)
	}
	if node.Expanded {
		for _, child := range node.Children {
			tw.loadExpanded(child)
		}
	}
}

// cursorNode returns the node currently under the tree's cursor.
// The Tree type does not export its cursor index, so we flatten the tree
// ourselves and use the scroll-offset heuristic together with ContentSize
// to determine the cursor position. However, an easier and fully correct
// approach is to simply mirror the same flatten logic the Tree uses and
// count from the top: the tree always keeps the cursor clamped, and we
// can observe it indirectly because the tree stores it as an index into
// the flat list. We store last-known cursor state ourselves.
//
// Since we cannot read tree.cursor directly, we track the cursor by
// wrapping HandleEvent (see above). For the initial state we just pick
// the first node.
func (tw *treeWrapper) cursorNode() *widget.TreeNode {
	// Flatten the tree the same way widget.Tree does.
	var flat []*widget.TreeNode
	for _, root := range tw.tree.Roots {
		flat = flatten(root, flat)
	}
	if len(flat) == 0 {
		return nil
	}

	// The tree clamps scrollY so that the cursor is always visible.
	// ScrollOffset gives us scrollY. The cursor is within
	// [scrollY, scrollY+viewHeight). For an initial best-guess we use
	// scrollY as the cursor if we have no better information. This gives
	// a reasonable default.
	_, scrollY := tw.tree.ScrollOffset()
	idx := scrollY
	if idx < 0 {
		idx = 0
	}
	if idx >= len(flat) {
		idx = len(flat) - 1
	}
	return flat[idx]
}

func flatten(node *widget.TreeNode, out []*widget.TreeNode) []*widget.TreeNode {
	if node == nil {
		return out
	}
	out = append(out, node)
	if node.Expanded {
		for _, child := range node.Children {
			out = flatten(child, out)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Filesystem helpers
// ---------------------------------------------------------------------------

// makeDirNode creates a TreeNode for a directory. A placeholder child is added
// so the tree shows an expand marker; real children are loaded lazily.
func (fv *fsViewer) makeDirNode(path string) *widget.TreeNode {
	info, _ := os.Lstat(path)
	name := filepath.Base(path)

	entry := &fsEntry{
		Path:    path,
		Info:    info,
		IsDir:   true,
		Hidden:  strings.HasPrefix(name, "."),
		Symlink: info != nil && info.Mode()&fs.ModeSymlink != 0,
	}
	if entry.Symlink {
		entry.LinkTgt, _ = os.Readlink(path)
	}

	return &widget.TreeNode{
		Label:    dirLabel(name),
		Data:     entry,
		Children: []*widget.TreeNode{{}}, // placeholder → non-leaf
	}
}

func dirLabel(name string) string  { return "\U0001F4C1 " + name }
func fileLabel(name string) string { return "   " + name }
func linkLabel(name string) string { return "\U0001F517 " + name }

// loadChildren reads a directory from disk and populates node.Children.
func (fv *fsViewer) loadChildren(node *widget.TreeNode) {
	entry, ok := node.Data.(*fsEntry)
	if !ok || !entry.IsDir {
		return
	}
	entry.Loaded = true

	dirEntries, err := os.ReadDir(entry.Path)
	if err != nil {
		node.Children = []*widget.TreeNode{{
			Label: "(access denied)",
			Data:  &fsEntry{Path: entry.Path},
		}}
		return
	}

	var dirs, files []*widget.TreeNode
	for _, de := range dirEntries {
		name := de.Name()
		hidden := strings.HasPrefix(name, ".")
		if !fv.showHidden && hidden {
			continue
		}

		childPath := filepath.Join(entry.Path, name)
		info, err := de.Info()
		if err != nil {
			continue
		}

		isSymlink := de.Type()&fs.ModeSymlink != 0
		isDir := de.IsDir()
		if isSymlink {
			if resolved, err := os.Stat(childPath); err == nil && resolved.IsDir() {
				isDir = true
			}
		}

		fe := &fsEntry{
			Path:    childPath,
			Info:    info,
			IsDir:   isDir,
			Hidden:  hidden,
			Symlink: isSymlink,
		}
		if isSymlink {
			fe.LinkTgt, _ = os.Readlink(childPath)
		}

		if isDir {
			label := dirLabel(name)
			if isSymlink {
				label = linkLabel(name)
			}
			tn := &widget.TreeNode{
				Label:    label,
				Data:     fe,
				Children: []*widget.TreeNode{{}}, // placeholder
			}
			dirs = append(dirs, tn)
		} else {
			label := fileLabel(name)
			if isSymlink {
				label = linkLabel(name)
			}
			tn := &widget.TreeNode{
				Label: label,
				Data:  fe,
			}
			files = append(files, tn)
		}
	}

	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(filepath.Base(dirs[i].Data.(*fsEntry).Path)) <
			strings.ToLower(filepath.Base(dirs[j].Data.(*fsEntry).Path))
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(filepath.Base(files[i].Data.(*fsEntry).Path)) <
			strings.ToLower(filepath.Base(files[j].Data.(*fsEntry).Path))
	})

	node.Children = append(dirs, files...)
	if len(node.Children) == 0 {
		node.Children = nil
	}
}

// ---------------------------------------------------------------------------
// Preview
// ---------------------------------------------------------------------------

func (fv *fsViewer) showPreview(node *widget.TreeNode) {
	entry, ok := node.Data.(*fsEntry)
	if !ok {
		return
	}

	fv.prevPanel.Title = filepath.Base(entry.Path)

	headerStyle := text.Style{}.WithFg(text.Yellow()).Bold()
	infoStyle := text.Style{}.WithFg(text.White())
	dimStyle := text.Style{}.Dim()

	var lines []text.StyledLine
	lines = append(lines, text.Styled("Path: "+entry.Path, infoStyle))
	lines = append(lines, text.StyledLine{})

	info, err := os.Lstat(entry.Path)
	if err != nil {
		lines = append(lines, text.Styled("Error: "+err.Error(), text.Style{}.WithFg(text.Red())))
		fv.preview.SetLines(lines)
		fv.preview.SetScrollOffset(0, 0)
		return
	}

	lines = append(lines, text.Styled(fmt.Sprintf("Permissions: %s", info.Mode()), infoStyle))
	lines = append(lines, text.Styled(fmt.Sprintf("Modified:    %s", info.ModTime().Format(time.RFC3339)), infoStyle))

	if info.Mode()&fs.ModeSymlink != 0 {
		target, err := os.Readlink(entry.Path)
		if err != nil {
			lines = append(lines, text.Styled("Symlink target: (error)", text.Style{}.WithFg(text.Red())))
		} else {
			lines = append(lines, text.Styled("Symlink target: "+target, text.Style{}.WithFg(text.Cyan())))
		}
		lines = append(lines, text.StyledLine{})
	}

	switch {
	case entry.IsDir:
		lines = append(lines, text.Styled("--- Directory ---", headerStyle))
		dirEntries, err := os.ReadDir(entry.Path)
		if err != nil {
			lines = append(lines, text.Styled("Cannot read: "+err.Error(), text.Style{}.WithFg(text.Red())))
		} else {
			nDirs, nFiles := 0, 0
			var totalSize int64
			for _, de := range dirEntries {
				if de.IsDir() {
					nDirs++
				} else {
					nFiles++
					if fi, err := de.Info(); err == nil {
						totalSize += fi.Size()
					}
				}
			}
			lines = append(lines, text.Styled(
				fmt.Sprintf("Items:       %d (%d dirs, %d files)", len(dirEntries), nDirs, nFiles), infoStyle))
			lines = append(lines, text.Styled(
				fmt.Sprintf("Total size:  %s", formatSize(totalSize)), infoStyle))
		}

	case info.Mode().IsRegular():
		lines = append(lines, text.Styled(
			fmt.Sprintf("Size:        %s (%d bytes)", formatSize(info.Size()), info.Size()), infoStyle))
		lines = append(lines, text.StyledLine{})

		if info.Size() == 0 {
			lines = append(lines, text.Styled("(empty file)", dimStyle))
		} else {
			f, err := os.Open(entry.Path)
			if err != nil {
				lines = append(lines, text.Styled("Cannot read: "+err.Error(), text.Style{}.WithFg(text.Red())))
			} else {
				buf := make([]byte, 8192)
				n, _ := f.Read(buf)
				f.Close()
				buf = buf[:n]

				probe := buf
				if len(probe) > 512 {
					probe = probe[:512]
				}

				if isBinary(probe) {
					lines = append(lines, text.Styled("--- Binary (hex) ---", headerStyle))
					for _, hl := range hexPreview(buf, 50) {
						lines = append(lines, text.Styled(hl, dimStyle))
					}
				} else {
					lines = append(lines, text.Styled("--- Content ---", headerStyle))
					textLines := strings.Split(string(buf), "\n")
					if len(textLines) > 50 {
						textLines = textLines[:50]
					}
					codeStyle := text.Style{}.WithFg(text.White())
					for i, tl := range textLines {
						num := fmt.Sprintf("%4d ", i+1)
						lines = append(lines, text.StyledLine{
							{Text: num, Style: dimStyle},
							{Text: tl, Style: codeStyle},
						})
					}
					if len(textLines) == 50 {
						lines = append(lines, text.Styled("  ... (truncated)", dimStyle))
					}
				}
			}
		}

	default:
		lines = append(lines, text.Styled(fmt.Sprintf("Type: %s", info.Mode().Type()), infoStyle))
	}

	fv.preview.SetLines(lines)
	fv.preview.SetScrollOffset(0, 0)
}

// ---------------------------------------------------------------------------
// Utility functions
// ---------------------------------------------------------------------------

func isBinary(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

func hexPreview(data []byte, maxLines int) []string {
	dump := hex.Dump(data)
	lines := strings.Split(dump, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, "... (truncated)")
	}
	return lines
}

func formatSize(size int64) string {
	switch {
	case size >= 1<<30:
		return fmt.Sprintf("%.1f GiB", float64(size)/float64(1<<30))
	case size >= 1<<20:
		return fmt.Sprintf("%.1f MiB", float64(size)/float64(1<<20))
	case size >= 1<<10:
		return fmt.Sprintf("%.1f KiB", float64(size)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func (fv *fsViewer) toggleHidden() {
	fv.showHidden = !fv.showHidden
	fv.refreshTree()
	status := "shown"
	if !fv.showHidden {
		status = "hidden"
	}
	fv.statusBar.SetText(text.Plain(fmt.Sprintf(
		" %s | Hidden files: %s | h:toggle | q:quit | r:refresh", fv.rootPath, status)))
}

func (fv *fsViewer) refreshTree() {
	rootNode := fv.makeDirNode(fv.rootPath)
	rootNode.Expanded = true
	fv.loadChildren(rootNode)
	fv.tree.Roots = []*widget.TreeNode{rootNode}
	fv.app.RequestRender()
}

// ---------------------------------------------------------------------------
// main
// ---------------------------------------------------------------------------

func main() {
	rootPath := "."
	if len(os.Args) > 1 {
		rootPath = os.Args[1]
	}
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fsview: %v\n", err)
		os.Exit(1)
	}

	fv := &fsViewer{
		rootPath:   absRoot,
		showHidden: true,
	}

	// Build the initial tree.
	rootNode := fv.makeDirNode(absRoot)
	rootNode.Expanded = true
	fv.loadChildren(rootNode)

	fv.tree = widget.NewTree([]*widget.TreeNode{rootNode})
	fv.wrapper = &treeWrapper{fv: fv, tree: fv.tree}

	// OnSelect fires for leaf (file) nodes on Enter.
	fv.tree.OnSelect = func(node *widget.TreeNode) {
		fv.showPreview(node)
	}

	fv.treePanel = widget.NewPanel(absRoot, fv.wrapper)
	fv.treePanel.Border = widget.BorderRounded

	fv.preview = widget.NewTextView()
	fv.preview.AutoScroll = false
	fv.prevPanel = widget.NewPanel("Preview", fv.preview)
	fv.prevPanel.Border = widget.BorderRounded

	fv.statusBar = &statusWidget{}
	fv.statusBar.SetText(text.Plain(fmt.Sprintf(
		" %s | h:toggle hidden | q:quit | r:refresh | Tab:switch focus", absRoot)))

	hsplit := layout.NewHSplit(
		layout.SplitChild{Renderable: fv.treePanel, Size: 0.35},
		layout.SplitChild{Renderable: fv.prevPanel, Size: 0.65},
	)
	root := layout.NewVSplit(
		layout.SplitChild{Renderable: hsplit, Size: 0.99},
		layout.SplitChild{Renderable: fv.statusBar, Size: 2},
	)

	t, err := term.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fsview: %v\n", err)
		os.Exit(1)
	}
	defer t.Close()

	fv.app = layout.NewApp(t, root)
	fv.app.AddFocusable(fv.wrapper)
	fv.app.SetFocus(fv.wrapper)
	fv.treePanel.Focused = true

	// Global key bindings.
	fv.app.BindRune('q', 0, func() { fv.app.Quit() })
	fv.app.BindRune('h', 0, func() { fv.toggleHidden() })
	fv.app.BindRune('r', 0, func() { fv.refreshTree() })

	// Show preview for the root directory initially.
	fv.showPreview(rootNode)

	if err := fv.app.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "fsview: %v\n", err)
		os.Exit(1)
	}
}
