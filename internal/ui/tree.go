package ui

import (
	"ShortcutVisualizer/internal/icon"
	"ShortcutVisualizer/internal/utils"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type TreeView struct {
	tree      *widget.Tree
	shortcuts map[string]string
	nodes     map[string][]string
	iconCache *icon.Cache
}

func NewTreeView() *TreeView {
	return &TreeView{
		shortcuts: make(map[string]string),
		nodes:     make(map[string][]string),
		iconCache: icon.NewCache(),
	}
}

func (w *MainWindow) childUIDs(uid string) []string {
	if uid == "" {
		return []string{"root"}
	}
	return w.tree.nodes[uid]
}

func (w *MainWindow) isBranch(uid string) bool {
	if uid == "" || uid == "root" {
		return true
	}
	path := w.tree.shortcuts[uid]
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
			label.SetText(filepath.Base(w.tree.shortcuts[uid]))
		}
	} else {
		absPath, _ := filepath.Abs(w.tree.shortcuts[uid])
		icon.SetResource(w.tree.iconCache.GetIconResource(absPath)) // 使用缓存机制获取图标
		//icon.SetResource(test1Res)
		//icon.SetResource(theme.FileIcon())
		// 对快捷方式文件名进行清理
		filename := filepath.Base(w.tree.shortcuts[uid])
		cleanName := utils.CleanShortcutName(filename)
		label.SetText(cleanName)
	}
}

func (w *MainWindow) onDoubleClick(uid string) {
	if path, ok := w.tree.shortcuts[uid]; ok {
		if utils.IsShortcut(path) {
			// 如果是快捷方式，则打开它
			w.openShortcut(path)
		} else {
			// 如果是文件夹，则切换其展开/折叠状态
			if w.tree.tree.IsBranch(uid) {
				if w.tree.tree.IsBranchOpen(uid) {
					w.tree.tree.CloseBranch(uid)
				} else {
					w.tree.tree.OpenBranch(uid)
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
	w.tree.iconCache.Clear() // 新增：清理图标缓存
	w.tree.shortcuts = make(map[string]string)
	w.tree.nodes = make(map[string][]string)

	// 从根目录开始遍历
	rootID := "root"
	w.tree.shortcuts[rootID] = folderPath
	w.traverseFolder(folderPath, rootID)

	// 刷新树形控件
	w.tree.tree.Refresh()

	// 自动展开根节点
	w.tree.tree.OpenBranch("root")
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
			w.tree.shortcuts[nodeID] = entryPath
			children = append(children, nodeID)
			w.traverseFolder(entryPath, nodeID)
		} else if strings.HasSuffix(strings.ToLower(entry.Name()), ".lnk") {
			// 处理快捷方式文件
			w.tree.shortcuts[nodeID] = entryPath
			children = append(children, nodeID)
		}
	}

	// 保存父节点与子节点的关系
	w.tree.nodes[parentID] = children
}
