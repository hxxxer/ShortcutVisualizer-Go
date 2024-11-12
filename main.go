//go:build windows

package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

type MainWindow struct {
	window    fyne.Window
	tree      *widget.Tree
	pathEntry *widget.Entry
	browseBtn *widget.Button
	shortcuts map[string]string   // 存储路径映射
	nodes     map[string][]string // 存储节点层级关系
}

func NewMainWindow(app fyne.App) *MainWindow {
	w := &MainWindow{
		window:    app.NewWindow("打开软件"),
		shortcuts: make(map[string]string),
		nodes:     make(map[string][]string),
	}

	w.createUI()
	return w
}

func (w *MainWindow) createUI() {
	// 创建路径输入框
	w.pathEntry = widget.NewEntry()
	w.pathEntry.SetText("D:\\软件")

	// 创建浏览按钮
	w.browseBtn = widget.NewButton("浏览", w.browseFolderDialog)

	// 创建树形控件
	w.tree = widget.NewTree(
		w.childUIDs,
		w.isBranch,
		w.createNode,
		w.updateNode,
	)

	// 设置双击处理
	w.tree.OnSelected = w.onDoubleClick

	// 创建顶部工具栏布局
	toolBar := container.NewBorder(nil, nil, nil, w.browseBtn, w.pathEntry)

	// 主布局
	content := container.NewBorder(toolBar, nil, nil, nil, w.tree)

	w.window.SetContent(content)

	// 初始加载文件夹
	w.populateTree(w.pathEntry.Text)
}

func (w *MainWindow) browseFolderDialog() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, w.window)
			return
		}
		if uri == nil {
			return
		}

		path := uri.Path()
		w.pathEntry.SetText(path)
		w.populateTree(path)
	}, w.window)
}

func (w *MainWindow) childUIDs(uid string) []string {
	if uid == "" {
		return []string{"root"}
	}
	return w.nodes[uid]
}

func (w *MainWindow) isBranch(uid string) bool {
	if uid == "" || uid == "root" {
		return true
	}
	path := w.shortcuts[uid]
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (w *MainWindow) createNode(branch bool) fyne.CanvasObject {
	return container.NewHBox(
		widget.NewIcon(theme.FolderIcon()),
		widget.NewLabel(""),
	)
}

func (w *MainWindow) updateNode(uid string, branch bool, node fyne.CanvasObject) {
	container := node.(*fyne.Container)
	icon := container.Objects[0].(*widget.Icon)
	label := container.Objects[1].(*widget.Label)

	if branch {
		icon.SetResource(theme.FolderIcon())
		if uid == "root" {
			label.SetText(filepath.Base(w.pathEntry.Text))
		} else {
			label.SetText(filepath.Base(w.shortcuts[uid]))
		}
	} else {
		icon.SetResource(theme.FileIcon())
		// 对快捷方式文件名进行清理
		filename := filepath.Base(w.shortcuts[uid])
		cleanName := cleanShortcutName(filename)
		label.SetText(cleanName)
	}
}

// 添加一个新的辅助函数来处理文件名
func cleanShortcutName(filename string) string {
	// 移除 .lnk 后缀
	name := strings.TrimSuffix(filename, ".lnk")

	// 使用正则表达式移除 ".exe - 快捷方式" 模式
	re := regexp.MustCompile(`(.+)\.exe.*$`)
	if match := re.FindStringSubmatch(name); len(match) > 1 {
		name = match[1]
	}

	return name
}

func (w *MainWindow) onSelected(uid string) {
	// 仅用于选择，双击处理在onDoubleClick中
}

func (w *MainWindow) onDoubleClick(uid string) {
	if path, ok := w.shortcuts[uid]; ok {
		if strings.HasSuffix(strings.ToLower(path), ".lnk") {
			// 如果是快捷方式，则打开它
			w.openShortcut(path)
		} else {
			// 如果是文件夹，则切换其展开/折叠状态
			if w.tree.IsBranch(uid) {
				if w.tree.IsBranchOpen(uid) {
					w.tree.CloseBranch(uid)
				} else {
					w.tree.OpenBranch(uid)
				}
			}
		}
	}
}

func (w *MainWindow) openShortcut(path string) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("explorer", path)
	} else {
		// 在其他系统上可能需要不同的处理方式
		return
	}

	err := cmd.Start()
	if err != nil {
		dialog.ShowError(err, w.window)
	}
}

func (w *MainWindow) populateTree(folderPath string) {
	// 清理原有数据
	w.shortcuts = make(map[string]string)
	w.nodes = make(map[string][]string)

	// 从根目录开始遍历
	rootID := "root"
	w.shortcuts[rootID] = folderPath
	w.traverseFolder(folderPath, rootID)

	// 刷新树形控件
	w.tree.Refresh()

	// 自动展开根节点
	w.tree.OpenBranch("root")
}

func (w *MainWindow) traverseFolder(path string, parentID string) {
	// 读取目录内容
	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}

	var children []string

	// 遍历目录项
	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		nodeID := filepath.Join(parentID, entry.Name())

		if entry.IsDir() {
			// 处理子目录
			w.shortcuts[nodeID] = entryPath
			children = append(children, nodeID)
			w.traverseFolder(entryPath, nodeID)
		} else if strings.HasSuffix(strings.ToLower(entry.Name()), ".lnk") {
			// 处理快捷方式文件
			w.shortcuts[nodeID] = entryPath
			children = append(children, nodeID)
		}
	}

	// 保存父节点与子节点的关系
	w.nodes[parentID] = children
}

func main() {
	a := app.New()
	w := NewMainWindow(a)

	w.window.Resize(fyne.NewSize(600, 600))
	w.window.ShowAndRun()
}
