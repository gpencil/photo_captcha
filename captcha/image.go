package captcha

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"net/http"
	"os"
	"strings"
	"time"
)

// BackgroundURLs 背景图列表（支持本地文件路径）
var BackgroundURLs = []string{
	//"images/image1.jpg",
	//"images/image2.jpg",
	"https://lunalab-res.oss-cn-hangzhou.aliyuncs.com/ttsVoice/captcha/image1.jpg",
	"https://lunalab-res.oss-cn-hangzhou.aliyuncs.com/ttsVoice/captcha/image2.jpg",
	"https://lunalab-res.oss-cn-hangzhou.aliyuncs.com/ttsVoice/captcha/image3.jpg",
	"https://lunalab-res.oss-cn-hangzhou.aliyuncs.com/ttsVoice/captcha/image4.jpg",
	"https://lunalab-res.oss-cn-hangzhou.aliyuncs.com/ttsVoice/captcha/image5.jpg",
	"https://lunalab-res.oss-cn-hangzhou.aliyuncs.com/ttsVoice/captcha/image6.jpg",
	"https://lunalab-res.oss-cn-hangzhou.aliyuncs.com/ttsVoice/captcha/image7.jpg",
	"https://lunalab-res.oss-cn-hangzhou.aliyuncs.com/ttsVoice/captcha/image8.jpg",
	"https://lunalab-res.oss-cn-hangzhou.aliyuncs.com/ttsVoice/captcha/image9.jpg",
	"https://lunalab-res.oss-cn-hangzhou.aliyuncs.com/ttsVoice/captcha/image10.jpg",
}

// DownloadImage 下载或加载图片（支持本地文件和网络URL）
func DownloadImage(pathOrURL string) (image.Image, error) {
	// 判断是本地文件还是网络URL
	if strings.HasPrefix(pathOrURL, "http://") || strings.HasPrefix(pathOrURL, "https://") {
		// 网络图片
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		resp, err := client.Get(pathOrURL)
		if err != nil {
			return nil, fmt.Errorf("failed to download image: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		img, _, err := image.Decode(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to decode image: %w", err)
		}

		return img, nil
	} else {
		// 本地文件
		file, err := os.Open(pathOrURL)
		if err != nil {
			return nil, fmt.Errorf("failed to open image file: %w", err)
		}
		defer file.Close()

		img, _, err := image.Decode(file)
		if err != nil {
			return nil, fmt.Errorf("failed to decode image: %w", err)
		}

		return img, nil
	}
}

// ImageToBase64 将图片转换为base64字符串
func ImageToBase64(img image.Image, format string) (string, error) {
	var buf []byte
	var err error

	switch format {
	case "png":
		buf, err = encodePNG(img)
	case "jpeg", "jpg":
		buf, err = encodeJPEG(img)
	default:
		buf, err = encodePNG(img)
		format = "png"
	}

	if err != nil {
		return "", err
	}

	base64Str := base64.StdEncoding.EncodeToString(buf)
	mimeType := "image/png"
	if format == "jpeg" || format == "jpg" {
		mimeType = "image/jpeg"
	}

	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Str), nil
}

// encodePNG 编码为PNG格式
func encodePNG(img image.Image) ([]byte, error) {
	buf := make([]byte, 0)
	w := &writerBuffer{buf: buf}

	encoder := png.Encoder{CompressionLevel: png.DefaultCompression}
	err := encoder.Encode(w, img)

	return w.buf, err
}

// encodeJPEG 编码为JPEG格式
func encodeJPEG(img image.Image) ([]byte, error) {
	buf := make([]byte, 0)
	w := &writerBuffer{buf: buf}

	err := jpeg.Encode(w, img, &jpeg.Options{Quality: 90})

	return w.buf, err
}

// writerBuffer 实现io.Writer接口的缓冲区
type writerBuffer struct {
	buf []byte
}

func (w *writerBuffer) Write(p []byte) (n int, err error) {
	w.buf = append(w.buf, p...)
	return len(p), nil
}

// GenerateCaptchaImages 生成验证码图片
func GenerateCaptchaImages(bgImage image.Image, x, y int, shape *PuzzleShape) (bgWithHole string, sliderPiece string, err error) {
	// 先将图片缩放到目标尺寸（350x200）
	targetWidth := 350
	targetHeight := 200
	resizedImage := ResizeImage(bgImage, targetWidth, targetHeight)

	// 根据缩放比例调整缺口位置
	scaleX := float64(targetWidth) / float64(bgImage.Bounds().Dx())
	scaleY := float64(targetHeight) / float64(bgImage.Bounds().Dy())
	scaledX := int(float64(x) * scaleX)
	scaledY := int(float64(y) * scaleY)

	// 创建带缺口的背景图
	holeImage := CreatePuzzleHole(resizedImage, scaledX, scaledY, shape)

	// 提取拼图块
	pieceImage := ExtractPuzzlePiece(resizedImage, scaledX, scaledY, shape)

	// 转换为base64
	bgBase64, err := ImageToBase64(holeImage, "png")
	if err != nil {
		return "", "", fmt.Errorf("failed to encode background: %w", err)
	}

	sliderBase64, err := ImageToBase64(pieceImage, "png")
	if err != nil {
		return "", "", fmt.Errorf("failed to encode slider: %w", err)
	}

	return bgBase64, sliderBase64, nil
}

// ResizeImage 缩放图片到指定尺寸（使用双线性插值，更平滑）
func ResizeImage(src image.Image, width, height int) image.Image {
	// 创建目标尺寸的图像
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// 使用双线性插值进行缩放
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// 计算源图像中的对应位置（浮点坐标）
			srcX := float64(x) * float64(srcW) / float64(width)
			srcY := float64(y) * float64(srcH) / float64(height)

			// 双线性插值
			x0 := int(srcX)
			y0 := int(srcY)
			x1 := x0 + 1
			y1 := y0 + 1

			// 边界检查
			if x1 >= srcW {
				x1 = srcW - 1
			}
			if y1 >= srcH {
				y1 = srcH - 1
			}

			// 获取四个邻近像素
			c00 := src.At(x0, y0)
			c01 := src.At(x0, y1)
			c10 := src.At(x1, y0)
			c11 := src.At(x1, y1)

			// 计算插值权重
			fx := srcX - float64(x0)
			fy := srcY - float64(y0)

			// 双线性插值混合
			r00, g00, b00, a00 := c00.RGBA()
			r01, g01, b01, a01 := c01.RGBA()
			r10, g10, b10, a10 := c10.RGBA()
			r11, g11, b11, a11 := c11.RGBA()

			// 混合权重 (0-65535)
			wx := uint32(fx * 65535)
			wy := uint32(fy * 65535)
			wx_inv := 65535 - wx
			wy_inv := 65535 - wy

			// 双线性插值
			r := (r00*wx_inv+r10*wx)/65535*wy_inv/65535 + (r01*wx_inv+r11*wx)/65535*wy/65535
			g := (g00*wx_inv+g10*wx)/65535*wy_inv/65535 + (g01*wx_inv+g11*wx)/65535*wy/65535
			b := (b00*wx_inv+b10*wx)/65535*wy_inv/65535 + (b01*wx_inv+b11*wx)/65535*wy/65535
			a := (a00*wx_inv+a10*wx)/65535*wy_inv/65535 + (a01*wx_inv+a11*wx)/65535*wy/65535

			// 转换为8位并设置像素
			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
	}

	return dst
}
