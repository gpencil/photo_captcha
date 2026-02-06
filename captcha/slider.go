package captcha

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// PuzzleShape 拼图形状参数
type PuzzleShape struct {
	Type        PuzzleType // 形状类型
	RightTab    bool       // 右侧是否有凸起（仅经典拼图使用）
	RightRadius int        // 右侧凸起半径
	RightY      int        // 右侧凸起Y位置

	LeftTab    bool // 左侧是否有凹槽（仅经典拼图使用）
	LeftRadius int
	LeftY      int

	TopTab    bool // 顶部是否有凸起（仅经典拼图使用）
	TopRadius int
	TopX      int

	BottomTab    bool // 底部是否有凹槽（仅经典拼图使用）
	BottomRadius int
	BottomX      int
}

// GenerateRandomPuzzleShape 生成随机拼图形状
func GenerateRandomPuzzleShape() *PuzzleShape {
	// 随机选择mask目录下存在的图形
	shapeType := PuzzleType(rand.Intn(4)) // 0-3 共4种形状

	// 打印日志
	var shapeName string
	switch shapeType {
	case PuzzleTypeTriangle:
		shapeName = "三角形"
	case PuzzleTypeHexagon:
		shapeName = "六边形"
	case PuzzleTypeTrapezoid:
		shapeName = "梯形"
	case PuzzleTypeStar:
		shapeName = "星形"
	}
	fmt.Printf("[生成的图形] %s (Type=%d)\n", shapeName, shapeType)

	return &PuzzleShape{
		Type: shapeType,
	}
}

// PuzzleType 拼图形状类型
type PuzzleType int

const (
	PuzzleTypeTriangle  PuzzleType = iota // 三角形
	PuzzleTypeHexagon                     // 六边形
	PuzzleTypeTrapezoid                   // 梯形
	PuzzleTypeStar                        // 星形
)

// PuzzleShape 拼图形状参数
type SliderCaptcha struct {
	ID         string `json:"id"`
	Background string `json:"background"` // 背景图base64
	Slider     string `json:"slider"`     // 滑块图base64
	PositionY  int    `json:"positionY"`  // 滑块Y轴位置
}

// Generate 生成新的滑块验证码
func Generate() (*SliderCaptcha, error) {
	// 随机选择背景图URL
	rand.Seed(time.Now().UnixNano())
	bgIndex := rand.Intn(len(BackgroundURLs))
	bgURL := BackgroundURLs[bgIndex]

	// 下载背景图
	bgImage, err := DownloadImage(bgURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download background image: %w", err)
	}

	// 获取图片尺寸
	bounds := bgImage.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// 随机生成缺口位置
	// X坐标: 在图片中间垂直线左右浮动
	centerX := imgWidth / 2
	// 在中间线附近浮动，范围为中间线的 ±25%
	offsetRangeX := int(float64(imgWidth) * 0.25)
	minX := centerX - offsetRangeX
	maxX := centerX + offsetRangeX - PuzzleWidth

	// 确保不超出边界
	if minX < 0 {
		minX = 0
	}
	if maxX > imgWidth-PuzzleWidth {
		maxX = imgWidth - PuzzleWidth
	}
	if maxX < minX {
		maxX = minX + PuzzleWidth
	}

	positionX := rand.Intn(maxX-minX) + minX

	// Y坐标: 在图片中间水平线上下浮动
	centerY := imgHeight / 2
	// 在中间线附近浮动，范围为中间线的 ±15%
	offsetRangeY := int(float64(imgHeight) * 0.15)
	minY := centerY - offsetRangeY
	maxY := centerY + offsetRangeY - PuzzleHeight

	// 确保不超出边界
	if minY < 0 {
		minY = 0
	}
	if maxY > imgHeight-PuzzleHeight {
		maxY = imgHeight - PuzzleHeight
	}
	if maxY < minY {
		maxY = minY + PuzzleHeight
	}

	positionY := rand.Intn(maxY-minY) + minY

	// 生成随机拼图形状参数
	puzzleShape := GenerateRandomPuzzleShape()

	// 生成验证码图片（内部会进行缩放）
	bgWithHole, sliderPiece, err := GenerateCaptchaImages(bgImage, positionX, positionY, puzzleShape)
	if err != nil {
		return nil, fmt.Errorf("failed to generate captcha images: %w", err)
	}

	// 计算缩放后的坐标（用于前端显示）
	targetWidth := 350
	targetHeight := 200
	scaleX := float64(targetWidth) / float64(imgWidth)
	scaleY := float64(targetHeight) / float64(imgHeight)
	scaledPositionX := int(float64(positionX) * scaleX)
	scaledPositionY := int(float64(positionY) * scaleY)

	// 生成唯一ID
	id := uuid.New().String()

	// 存储验证码数据（使用原始坐标用于验证）
	captchaData := &CaptchaData{
		ID:        id,
		PositionX: scaledPositionX, // 使用缩放后的坐标
		PositionY: scaledPositionY,
	}
	Set(id, captchaData)

	return &SliderCaptcha{
		ID:         id,
		Background: bgWithHole,
		Slider:     sliderPiece,
		PositionY:  scaledPositionY, // 返回缩放后的Y坐标
	}, nil
}

// Verify 验证滑块位置
// tolerance: 允许的误差范围（像素）
func Verify(id string, userX int, tolerance int) (bool, error) {
	// 获取存储的验证码数据
	data, exists := Get(id)
	if !exists {
		return false, fmt.Errorf("captcha not found or expired")
	}

	// 计算误差
	diff := abs(userX - data.PositionX)

	// 验证是否在误差范围内
	success := diff <= tolerance

	// 验证后删除验证码（无论成功还是失败）
	if success {
		Delete(id)
	}

	return success, nil
}

// abs 返回绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// VerifyWithTolerance 使用默认误差(5像素)验证
func VerifyWithTolerance(id string, userX int) (bool, error) {
	return Verify(id, userX, 5)
}
