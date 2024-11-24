package icon

import (
	"bytes"
	"fyne.io/fyne/v2"
	"image"
	"image/png"
)

// ImageToResource 将image.Image转换为fyne.Resource
func ImageToResource(img image.Image) (fyne.Resource, error) {
	// 创建一个buffer来存储PNG数据
	var buf bytes.Buffer

	// 将image编码为PNG格式
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}

	// 创建fyne的StaticResource
	// 使用当前时间戳作为唯一资源名
	return fyne.NewStaticResource("icon", buf.Bytes()), nil
}
