package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type MainWindow struct {
	window    fyne.Window
	tree      *TreeView // 移到单独的类型
	pathEntry *widget.Entry
	browseBtn *widget.Button
}

func NewMainWindow(app fyne.App) *MainWindow {
	w := &MainWindow{
		window: app.NewWindow("打开软件"),
		tree:   NewTreeView(),
	}
	w.createUI()
	return w
}

func (w *MainWindow) Window() fyne.Window {
	return w.window
}

func (w *MainWindow) createUI() {
	// 创建路径输入框
	w.pathEntry = widget.NewEntry()
	w.pathEntry.SetText("D:\\软件")

	// 创建浏览按钮
	w.browseBtn = widget.NewButton("浏览", w.browseFolderDialog)

	// 创建树形控件
	w.tree.tree = widget.NewTree(
		w.childUIDs,
		w.isBranch,
		w.createNode,
		w.updateNode,
	)

	// 设置双击处理
	w.tree.tree.OnSelected = w.onDoubleClick

	// 创建顶部工具栏布局
	toolBar := container.NewBorder(nil, nil, nil, w.browseBtn, w.pathEntry)

	// 主布局
	content := container.NewBorder(toolBar, nil, nil, nil, w.tree.tree)

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
