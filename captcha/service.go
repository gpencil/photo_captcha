package captcha

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CaptchaService 验证码服务（预加载优化版）
type CaptchaService struct {
	// 预加载的背景图片
	backgroundImages []image.Image
	// 预生成的拼图mask
	puzzleMasks map[PuzzleType]*image.Alpha
	// 背景图片URL列表（OSS或本地）
	backgroundURLs []string
	// 读写锁
	mu sync.RWMutex
	// 是否已初始化
	initialized bool
}

// NewCaptchaService 创建验证码服务实例
func NewCaptchaService() *CaptchaService {
	return &CaptchaService{
		backgroundImages: make([]image.Image, 0),
		puzzleMasks:      make(map[PuzzleType]*image.Alpha),
		backgroundURLs:   make([]string, 0),
	}
}

// SetBackgroundURLs 设置背景图片URL列表
func (s *CaptchaService) SetBackgroundURLs(urls []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.backgroundURLs = urls
}

// Init 初始化验证码服务（在服务启动时调用）
func (s *CaptchaService) Init() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.initialized {
		return nil
	}

	fmt.Println("[Captcha] 开始初始化验证码服务...")

	// 如果没有设置URL列表，使用全局配置
	if len(s.backgroundURLs) == 0 {
		s.backgroundURLs = BackgroundURLs
	}

	// 1. 从OSS/本地预加载所有背景图片（只下载一次）
	if err := s.loadBackgroundImages(); err != nil {
		return fmt.Errorf("加载背景图片失败: %w", err)
	}
	fmt.Printf("[Captcha] 成功加载并缓存 %d 张背景图片\n", len(s.backgroundImages))

	// 2. 预生成拼图mask
	if err := s.generatePuzzleMasks(); err != nil {
		return fmt.Errorf("生成拼图mask失败: %w", err)
	}
	fmt.Printf("[Captcha] 成功生成 %d 种拼图mask\n", len(s.puzzleMasks))

	s.initialized = true
	fmt.Println("[Captcha] 验证码服务初始化完成")

	return nil
}

// loadBackgroundImages 从OSS或本地预加载所有背景图片（只下载一次，缓存到内存）
func (s *CaptchaService) loadBackgroundImages() error {
	for i, imgURL := range s.backgroundURLs {
		// DownloadImage 会自动判断是本地文件还是OSS URL
		img, err := DownloadImage(imgURL)
		if err != nil {
			return fmt.Errorf("加载图片 %s 失败: %w", imgURL, err)
		}

		// 缓存到内存
		s.backgroundImages = append(s.backgroundImages, img)

		// 判断来源并输出日志
		source := "本地"
		if len(imgURL) > 4 && (imgURL[:4] == "http" || imgURL[:5] == "https") {
			source = "OSS"
		}

		fmt.Printf("[Captcha]   - 从%s加载并缓存图片 %d: %s (%dx%d)\n",
			source, i+1, imgURL, img.Bounds().Dx(), img.Bounds().Dy())
	}
	return nil
}

// generatePuzzleMasks 预生成所有拼图mask
func (s *CaptchaService) generatePuzzleMasks() error {
	shapeTypes := []PuzzleType{
		PuzzleTypeTriangle,
		PuzzleTypeHexagon,
		PuzzleTypeTrapezoid,
		PuzzleTypeStar,
	}

	for _, shapeType := range shapeTypes {
		shape := &PuzzleShape{Type: shapeType}
		mask := GeneratePuzzleMask(shape)
		s.puzzleMasks[shapeType] = mask

		shapeName := getShapeName(shapeType)
		fmt.Printf("[Captcha]   - 生成 %s mask\n", shapeName)
	}

	return nil
}

// GetRandomBackground 随机获取一个预加载的背景图片
func (s *CaptchaService) GetRandomBackground() image.Image {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.backgroundImages) == 0 {
		return nil
	}

	// 随机选择一个背景图片
	index := rand.Intn(len(s.backgroundImages))
	return s.backgroundImages[index]
}

// GetPuzzleMask 获取预生成的拼图mask
func (s *CaptchaService) GetPuzzleMask(shapeType PuzzleType) *image.Alpha {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.puzzleMasks[shapeType]
}

// Generate 生成验证码（使用预加载的资源）
func (s *CaptchaService) Generate() (*SliderCaptcha, error) {
	if !s.initialized {
		return nil, fmt.Errorf("captcha service not initialized, call Init() first")
	}

	// 使用预加载的背景图片
	bgImage := s.GetRandomBackground()
	if bgImage == nil {
		return nil, fmt.Errorf("no background images available")
	}

	// 获取图片尺寸
	bounds := bgImage.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// 随机生成缺口位置（在中心区域）
	centerX := imgWidth / 2
	centerY := imgHeight / 2
	offsetRangeX := int(float64(imgWidth) * 0.25)
	offsetRangeY := int(float64(imgHeight) * 0.15)

	minX := centerX - offsetRangeX
	maxX := centerX + offsetRangeX - PuzzleWidth
	if minX < 0 {
		minX = 0
	}
	if maxX > imgWidth-PuzzleWidth {
		maxX = imgWidth - PuzzleWidth
	}
	if maxX < minX {
		maxX = minX + PuzzleWidth
	}

	minY := centerY - offsetRangeY
	maxY := centerY + offsetRangeY - PuzzleHeight
	if minY < 0 {
		minY = 0
	}
	if maxY > imgHeight-PuzzleHeight {
		maxY = imgHeight - PuzzleHeight
	}
	if maxY < minY {
		maxY = minY + PuzzleHeight
	}

	rand.Seed(TimeNow().UnixNano())
	positionX := rand.Intn(maxX-minX) + minX
	positionY := rand.Intn(maxY-minY) + minY

	// 随机选择拼图形状
	shapeType := PuzzleType(rand.Intn(4))

	// 获取预生成的mask
	mask := s.GetPuzzleMask(shapeType)
	if mask == nil {
		return nil, fmt.Errorf("mask not found for shape type %d", shapeType)
	}

	// 生成验证码图片
	bgWithHole, sliderPiece, err := GenerateCaptchaImagesWithMask(bgImage, positionX, positionY, mask)
	if err != nil {
		return nil, fmt.Errorf("failed to generate captcha images: %w", err)
	}

	// 计算缩放后的坐标
	targetWidth := 350
	targetHeight := 200
	scaleX := float64(targetWidth) / float64(imgWidth)
	scaleY := float64(targetHeight) / float64(imgHeight)
	scaledPositionX := int(float64(positionX) * scaleX)
	scaledPositionY := int(float64(positionY) * scaleY)

	// 生成唯一ID
	id := uuid.New().String()

	// 存储验证码数据
	captchaData := &CaptchaData{
		ID:        id,
		PositionX: scaledPositionX,
		PositionY: scaledPositionY,
	}
	Set(id, captchaData)

	shapeName := getShapeName(shapeType)
	fmt.Printf("[生成的图形] %s (Type=%d)\n", shapeName, shapeType)

	return &SliderCaptcha{
		ID:         id,
		Background: bgWithHole,
		Slider:     sliderPiece,
		PositionY:  scaledPositionY,
	}, nil
}

// GenerateCaptchaImagesWithMask 使用预生成的mask生成验证码图片
func GenerateCaptchaImagesWithMask(bgImage image.Image, x, y int, mask *image.Alpha) (bgWithHole string, sliderPiece string, err error) {
	// 缩放到目标尺寸
	targetWidth := 350
	targetHeight := 200
	resizedImage := ResizeImage(bgImage, targetWidth, targetHeight)

	// 根据缩放比例调整缺口位置
	scaleX := float64(targetWidth) / float64(bgImage.Bounds().Dx())
	scaleY := float64(targetHeight) / float64(bgImage.Bounds().Dy())
	scaledX := int(float64(x) * scaleX)
	scaledY := int(float64(y) * scaleY)

	// 创建带缺口的背景图
	holeImage := CreatePuzzleHoleWithMask(resizedImage, scaledX, scaledY, mask)

	// 提取拼图块
	pieceImage := ExtractPuzzlePieceWithMask(resizedImage, scaledX, scaledY, mask)

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

// CreatePuzzleHoleWithMask 使用预生成的mask创建缺口
func CreatePuzzleHoleWithMask(bgImage image.Image, x, y int, mask *image.Alpha) image.Image {
	result := image.NewRGBA(bgImage.Bounds())
	draw.Draw(result, result.Bounds(), bgImage, image.Point{}, draw.Src)

	for py := 0; py < PuzzleHeight; py++ {
		for px := 0; px < PuzzleWidth; px++ {
			targetX := x + px
			targetY := y + py

			if targetX < 0 || targetX >= result.Bounds().Dx() ||
				targetY < 0 || targetY >= result.Bounds().Dy() {
				continue
			}

			alpha := mask.AlphaAt(px, py).A
			if alpha > 0 {
				c := result.RGBAAt(targetX, targetY)
				result.SetRGBA(targetX, targetY, color.RGBA{
					R: uint8(float64(c.R)*0.5 + 255*0.5),
					G: uint8(float64(c.G)*0.6 + 255*0.4),
					B: uint8(float64(c.B)*0.6 + 255*0.4),
					A: 255,
				})
			}
		}
	}

	addHoleBorder(result, mask, x, y)
	applyGaussianBlurToHole(result, mask, x, y)

	return result
}

// ExtractPuzzlePieceWithMask 使用预生成的mask提取拼图块
func ExtractPuzzlePieceWithMask(bgImage image.Image, x, y int, mask *image.Alpha) image.Image {
	piece := image.NewRGBA(image.Rect(0, 0, PuzzleWidth, PuzzleHeight))
	draw.Draw(piece, piece.Bounds(), image.Transparent, image.Point{}, draw.Src)

	for py := 0; py < PuzzleHeight; py++ {
		for px := 0; px < PuzzleWidth; px++ {
			alpha := mask.AlphaAt(px, py).A
			if alpha > 0 {
				srcX := x + px
				srcY := y + py

				if srcX >= 0 && srcX < bgImage.Bounds().Dx() &&
					srcY >= 0 && srcY < bgImage.Bounds().Dy() {
					c := bgImage.At(srcX, srcY)
					piece.Set(px, py, c)
				}
			}
		}
	}

	addSimpleBorder(piece, mask)
	add3DEffect(piece, mask)
	applyGaussianBlur(piece, mask)

	return piece
}

// getShapeName 获取形状名称
func getShapeName(shapeType PuzzleType) string {
	switch shapeType {
	case PuzzleTypeTriangle:
		return "三角形"
	case PuzzleTypeHexagon:
		return "六边形"
	case PuzzleTypeTrapezoid:
		return "梯形"
	case PuzzleTypeStar:
		return "星形"
	default:
		return "未知"
	}
}

// TimeNow 获取当前时间（方便mock测试）
func TimeNow() time.Time {
	return time.Now()
}
