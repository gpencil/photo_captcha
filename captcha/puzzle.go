package captcha

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
)

// PuzzleSize 拼图块大小
const (
	PuzzleWidth  = 70 // 修改这里可调整拼图宽度
	PuzzleHeight = 70 // 修改这里可调整拼图高度
)

// GeneratePuzzleMask 生成拼图形状的mask（优先使用预制图片）
func GeneratePuzzleMask(shape *PuzzleShape) *image.Alpha {
	// 优先尝试从mask目录加载预制图片
	maskFile := getMaskFile(shape.Type)
	if maskFile != "" {
		mask, err := loadMaskFromFile(maskFile)
		if err == nil {
			return mask
		}
		// 如果加载失败，回退到程序生成
		fmt.Printf("Failed to load mask from %s, using generated mask: %v\n", maskFile, err)
	}

	// 程序生成mask（后备方案）
	mask := image.NewAlpha(image.Rect(0, 0, PuzzleWidth, PuzzleHeight))

	// 绘制拼图形状
	for y := 0; y < PuzzleHeight; y++ {
		for x := 0; x < PuzzleWidth; x++ {
			if isInsidePuzzle(x, y, shape) {
				mask.SetAlpha(x, y, color.Alpha{A: 255})
			} else {
				mask.SetAlpha(x, y, color.Alpha{A: 0})
			}
		}
	}

	return mask
}

// getMaskFile 根据形状类型获取mask文件路径
func getMaskFile(shapeType PuzzleType) string {
	switch shapeType {
	case PuzzleTypeTriangle:
		return "mask/triangle.png"
	case PuzzleTypeHexagon:
		return "mask/hexagon.png"
	case PuzzleTypeTrapezoid:
		return "mask/trapezoid.png"
	case PuzzleTypeStar:
		return "mask/star.png"
	default:
		return ""
	}
}

// loadMaskFromFile 从文件加载mask并缩放到目标尺寸
func loadMaskFromFile(filename string) (*image.Alpha, error) {
	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 解码图片
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// 缩放到目标尺寸
	resizedImg := ResizeImage(img, PuzzleWidth, PuzzleHeight)

	// 转换为Alpha mask，保留原始alpha值（保持抗锯齿效果）
	mask := image.NewAlpha(image.Rect(0, 0, PuzzleWidth, PuzzleHeight))
	for y := 0; y < PuzzleHeight; y++ {
		for x := 0; x < PuzzleWidth; x++ {
			c := resizedImg.At(x, y)
			_, _, _, a := c.RGBA()
			// 直接使用原始alpha值（0-65535转为0-255）
			alpha8 := uint8(a >> 8)
			mask.SetAlpha(x, y, color.Alpha{A: alpha8})
		}
	}

	return mask, nil
}

// isInsidePuzzle 判断点是否在拼图形状内
func isInsidePuzzle(x, y int, shape *PuzzleShape) bool {
	// 根据形状类型调用不同的判断函数
	switch shape.Type {
	case PuzzleTypeTriangle:
		return isInsideTriangle(x, y)
	case PuzzleTypeHexagon:
		return isInsideHexagon(x, y)
	case PuzzleTypeTrapezoid:
		return isInsideTrapezoid(x, y)
	case PuzzleTypeStar:
		return isInsideStar(x, y)
	default:
		return isInsideTriangle(x, y)
	}
}

// isInsideTriangle 三角形（等腰三角形，顶点朝上，尖锐）
func isInsideTriangle(x, y int) bool {
	centerX := PuzzleWidth / 2
	marginBottom := 8 // 下边距
	marginSide := 8   // 左右边距

	// 检查边界
	if x < marginSide || x >= PuzzleWidth-marginSide {
		return false
	}
	if y >= PuzzleHeight-marginBottom {
		return false
	}

	// 三角形高度（从顶部到下边距）
	height := float64(PuzzleHeight - marginBottom)

	// 当前y在三角形中的相对位置（0到1）
	relativeY := float64(y) / height

	// 顶点宽度为0（真正的尖锐顶点）
	topWidth := 0.0
	bottomWidth := float64(PuzzleWidth - 2*marginSide)

	// 根据y坐标计算当前宽度
	currentWidth := topWidth + (bottomWidth-topWidth)*relativeY

	// 计算x到中心的距离
	dx := float64(x - centerX)

	return math.Abs(dx) <= currentWidth/2
}

// isInsideHexagon 六边形（6条直边的平顶正六边形）
func isInsideHexagon(x, y int) bool {
	centerX := PuzzleWidth / 2
	centerY := PuzzleHeight / 2
	radius := float64(PuzzleWidth/2 - 10)

	// 将坐标转换为相对于中心的坐标
	px := float64(x - centerX)
	py := float64(y - centerY)

	// 平顶六边形的6个顶点（从右上开始，顺时针）
	// 平顶六边形的顶点角度: 0°, 60°, 120°, 180°, 240°, 300°
	vertices := []struct{ x, y float64 }{
		{radius, 0},                               // 右
		{radius / 2, radius * math.Sqrt(3) / 2},   // 右下
		{-radius / 2, radius * math.Sqrt(3) / 2},  // 左下
		{-radius, 0},                              // 左
		{-radius / 2, -radius * math.Sqrt(3) / 2}, // 左上
		{radius / 2, -radius * math.Sqrt(3) / 2},  // 右上
	}

	// 使用叉积法检查点是否在多边形内
	// 对于每条边，检查点是否在边的内侧
	inside := true
	for i := 0; i < 6; i++ {
		j := (i + 1) % 6
		// 边从 vertices[i] 到 vertices[j]
		// 计算边的向量
		edgeX := vertices[j].x - vertices[i].x
		edgeY := vertices[j].y - vertices[i].y

		// 计算从顶点到测试点的向量
		pointX := px - vertices[i].x
		pointY := py - vertices[i].y

		// 计算叉积 (2D cross product)
		cross := edgeX*pointY - edgeY*pointX

		// 对于逆时针定义的多边形，如果点在内部，所有叉积应该 >= 0（或全部 <= 0）
		if cross < 0 {
			inside = false
			break
		}
	}

	return inside
}

// isInsideTrapezoid 梯形（倒置版，上窄下宽的等腰梯形）
func isInsideTrapezoid(x, y int) bool {
	centerX := PuzzleWidth / 2
	centerY := PuzzleHeight / 2

	// 梯形参数
	height := float64(PuzzleHeight - 20)     // 梯形高度
	topWidth := float64(PuzzleWidth - 40)    // 上底宽度（较窄）
	bottomWidth := float64(PuzzleWidth - 20) // 下底宽度（更宽）

	dx := float64(x - centerX)
	dy := float64(y - centerY)

	// 检查是否在高度范围内
	if math.Abs(dy) > height/2 {
		return false
	}

	// 根据y坐标计算当前宽度
	// y从 centerY-height/2 到 centerY+height/2
	normalizedY := (dy + height/2) / height // 0到1
	currentWidth := topWidth + (bottomWidth-topWidth)*normalizedY

	return math.Abs(dx) <= currentWidth/2
}

// isInsideStar 五角星（后备方案，当mask文件加载失败时使用）
func isInsideStar(x, y int) bool {
	centerX := PuzzleWidth / 2
	centerY := PuzzleHeight / 2

	outerRadius := float64(PuzzleWidth/2 - 8) // 外半径
	innerRadius := outerRadius * 0.4          // 内半径（五角星的凹陷）

	dx := float64(x - centerX)
	dy := float64(y - centerY)
	dist := math.Sqrt(dx*dx + dy*dy)

	// 转换为极坐标
	angle := math.Atan2(dy, dx)
	if angle < 0 {
		angle += 2 * math.Pi
	}

	// 五角星有5个角，每个角间隔72度（2π/5）
	segmentAngle := 2 * math.Pi / 5 // 72度
	halfAngle := math.Pi / 5        // 36度（半扇区）

	// 归一化角度，使其从第一个角开始（-π/2）
	normalizedAngle := angle + math.Pi/2
	if normalizedAngle < 0 {
		normalizedAngle += 2 * math.Pi
	}
	if normalizedAngle >= 2*math.Pi {
		normalizedAngle -= 2 * math.Pi
	}

	// 计算在扇区内的位置
	segmentIndex := int(normalizedAngle / segmentAngle)
	angleInSegment := normalizedAngle - float64(segmentIndex)*segmentAngle

	// 根据角度位置计算最大距离
	centerOffset := math.Abs(angleInSegment - halfAngle)
	maxDist := innerRadius + (outerRadius-innerRadius)*(1-centerOffset/halfAngle)

	return dist <= maxDist
}

// CreatePuzzleHole 在背景图上创建拼图缺口
func CreatePuzzleHole(bgImage image.Image, x, y int, shape *PuzzleShape) image.Image {
	// 创建可编辑的图像副本
	result := image.NewRGBA(bgImage.Bounds())
	draw.Draw(result, result.Bounds(), bgImage, image.Point{}, draw.Src)

	mask := GeneratePuzzleMask(shape)

	// 在指定位置绘制缺口 - 添加白色遮罩
	for py := 0; py < PuzzleHeight; py++ {
		for px := 0; px < PuzzleWidth; px++ {
			targetX := x + px
			targetY := y + py

			// 检查边界
			if targetX < 0 || targetX >= result.Bounds().Dx() ||
				targetY < 0 || targetY >= result.Bounds().Dy() {
				continue
			}

			alpha := mask.AlphaAt(px, py).A
			if alpha > 0 {
				c := result.RGBAAt(targetX, targetY)
				// 白色遮罩：混合原图和白色（降低白色遮罩浓度，让背景图更明显）
				result.SetRGBA(targetX, targetY, color.RGBA{
					R: uint8(float64(c.R)*0.4 + 255*0.6), // 40%原图 + 60%白色（原来是40%+60%） lcq1
					G: uint8(float64(c.G)*0.6 + 255*0.4),
					B: uint8(float64(c.B)*0.6 + 255*0.4),
					A: 255,
				})
			}
		}
	}

	// 添加缺口边框
	addHoleBorder(result, mask, x, y)

	// 对缺口边缘应用高斯模糊，让边缘更平滑
	applyGaussianBlurToHole(result, mask, x, y)

	return result
}

// addHoleBorder 添加缺口边框
func addHoleBorder(result *image.RGBA, mask *image.Alpha, x, y int) {
	borderColor := color.RGBA{R: 0, G: 0, B: 0, A: 0} // 降低边框不透明度，0去掉边框 150更明显 80更淡 lcq1

	for py := 0; py < PuzzleHeight; py++ {
		for px := 0; px < PuzzleWidth; px++ {
			if mask.AlphaAt(px, py).A > 0 {
				// 检查是否在边缘
				if isHoleEdge(px, py, mask) {
					targetX := x + px
					targetY := y + py
					if targetX >= 0 && targetX < result.Bounds().Dx() && targetY >= 0 && targetY < result.Bounds().Dy() {
						// 在边缘添加黑色描边
						result.SetRGBA(targetX, targetY, borderColor)
					}
				}
			}
		}
	}
}

// isHoleEdge 检查像素是否在缺口边缘
func isHoleEdge(x, y int, mask *image.Alpha) bool {
	// 检查周围像素
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx := x + dx
			ny := y + dy
			if nx < 0 || nx >= PuzzleWidth || ny < 0 || ny >= PuzzleHeight {
				return true
			}
			if mask.AlphaAt(nx, ny).A == 0 {
				return true
			}
		}
	}
	return false
}

// ExtractPuzzlePiece 从背景图提取拼图块
func ExtractPuzzlePiece(bgImage image.Image, x, y int, shape *PuzzleShape) image.Image {
	mask := GeneratePuzzleMask(shape)

	// 创建拼图块图像
	piece := image.NewRGBA(image.Rect(0, 0, PuzzleWidth, PuzzleHeight))

	// 初始化为透明
	draw.Draw(piece, piece.Bounds(), image.Transparent, image.Point{}, draw.Src)

	// 使用mask提取拼图块 - 只复制原始像素，不做任何处理
	for py := 0; py < PuzzleHeight; py++ {
		for px := 0; px < PuzzleWidth; px++ {
			alpha := mask.AlphaAt(px, py).A
			if alpha > 0 {
				srcX := x + px
				srcY := y + py

				// 检查边界
				if srcX >= 0 && srcX < bgImage.Bounds().Dx() &&
					srcY >= 0 && srcY < bgImage.Bounds().Dy() {
					c := bgImage.At(srcX, srcY)
					// 直接复制原始颜色，不做任何调整
					piece.Set(px, py, c)
				}
			}
		}
	}

	// 添加简单边框
	addSimpleBorder(piece, mask)

	// 添加立体感效果（高光）
	add3DEffect(piece, mask)

	// 添加轻微高斯模糊，进一步平滑边缘
	applyGaussianBlur(piece, mask)

	return piece
}

// addSimpleBorder 添加白色边框（带增强抗锯齿）
func addSimpleBorder(piece *image.RGBA, mask *image.Alpha) {
	// 先绘制基础边框
	borderColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	for py := 0; py < PuzzleHeight; py++ {
		for px := 0; px < PuzzleWidth; px++ {
			if mask.AlphaAt(px, py).A > 0 {
				if isEdgeSimple(px, py, mask) {
					piece.SetRGBA(px, py, borderColor)
				}
			}
		}
	}

	// 进行抗锯齿处理
	antiAliasEdges(piece, mask)
}

// antiAliasEdges 对边缘进行抗锯齿处理（超强平滑版）
func antiAliasEdges(piece *image.RGBA, mask *image.Alpha) {
	// 第一遍：对边缘的非白色像素进行强力抗锯齿
	for py := 0; py < PuzzleHeight; py++ {
		for px := 0; px < PuzzleWidth; px++ {
			if mask.AlphaAt(px, py).A > 0 {
				// 检查是否在边缘
				transparentNeighbors := countTransparentNeighbors(px, py, mask)
				if transparentNeighbors > 0 {
					current := piece.RGBAAt(px, py)

					// 如果是纯白色边框，跳过
					if current.R == 255 && current.G == 255 && current.B == 255 {
						continue
					}

					// 收集周围非白色像素，扩大范围到3像素
					var sumR, sumG, sumB uint32
					var totalWeight float64

					for dy := -3; dy <= 3; dy++ {
						for dx := -3; dx <= 3; dx++ {
							if dx == 0 && dy == 0 {
								continue
							}
							nx := px + dx
							ny := py + dy
							if nx >= 0 && nx < PuzzleWidth && ny >= 0 && ny < PuzzleHeight {
								if mask.AlphaAt(nx, ny).A > 0 {
									c := piece.RGBAAt(nx, ny)
									// 跳过白色边框像素
									if !(c.R == 255 && c.G == 255 && c.B == 255) {
										// 距离加权，越近权重越高
										distance := math.Sqrt(float64(dx*dx + dy*dy))
										weight := 1.0 / math.Pow(distance+1.0, 1.5) // 使用更强的衰减

										sumR += uint32(float64(c.R) * weight)
										sumG += uint32(float64(c.G) * weight)
										sumB += uint32(float64(c.B) * weight)
										totalWeight += weight
									}
								}
							}
						}
					}

					if totalWeight > 0 {
						// 计算加权平均
						avgR := sumR / uint32(totalWeight)
						avgG := sumG / uint32(totalWeight)
						avgB := sumB / uint32(totalWeight)

						// 根据边缘位置决定混合比例，提高到50%-90%
						mixRatio := 0.5 + float64(transparentNeighbors)/9.0*0.4

						piece.SetRGBA(px, py, color.RGBA{
							R: uint8(float64(current.R)*(1-mixRatio) + float64(avgR)*mixRatio),
							G: uint8(float64(current.G)*(1-mixRatio) + float64(avgG)*mixRatio),
							B: uint8(float64(current.B)*(1-mixRatio) + float64(avgB)*mixRatio),
							A: 255,
						})
					}
				}
			}
		}
	}

	// 第二遍：对斜边进行额外平滑（针对梯形）
	smoothDiagonalEdges(piece, mask)

	// 第三遍：全局轻微平滑，消除残留的锯齿
	globalSmooth(piece, mask)
}

// globalSmooth 对所有非边框像素进行轻微的全局平滑
func globalSmooth(piece *image.RGBA, mask *image.Alpha) {
	for py := 1; py < PuzzleHeight-1; py++ {
		for px := 1; px < PuzzleWidth-1; px++ {
			if mask.AlphaAt(px, py).A > 0 {
				current := piece.RGBAAt(px, py)

				// 跳过白色边框
				if current.R == 255 && current.G == 255 && current.B == 255 {
					continue
				}

				// 收集周围像素进行轻微平滑
				var sumR, sumG, sumB uint32
				var count uint32

				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						nx := px + dx
						ny := py + dy
						if mask.AlphaAt(nx, ny).A > 0 {
							c := piece.RGBAAt(nx, ny)
							if !(c.R == 255 && c.G == 255 && c.B == 255) {
								sumR += uint32(c.R)
								sumG += uint32(c.G)
								sumB += uint32(c.B)
								count++
							}
						}
					}
				}

				if count > 0 {
					avgR := sumR / count
					avgG := sumG / count
					avgB := sumB / count

					// 只做轻微平滑（20%混合）
					piece.SetRGBA(px, py, color.RGBA{
						R: uint8(float64(current.R)*0.8 + float64(avgR)*0.2),
						G: uint8(float64(current.G)*0.8 + float64(avgG)*0.2),
						B: uint8(float64(current.B)*0.8 + float64(avgB)*0.2),
						A: 255,
					})
				}
			}
		}
	}
}

// smoothDiagonalEdges 对斜边进行额外的平滑处理
func smoothDiagonalEdges(piece *image.RGBA, mask *image.Alpha) {
	for py := 1; py < PuzzleHeight-1; py++ {
		for px := 1; px < PuzzleWidth-1; px++ {
			if mask.AlphaAt(px, py).A > 0 {
				current := piece.RGBAAt(px, py)

				// 跳过白色边框
				if current.R == 255 && current.G == 255 && current.B == 255 {
					continue
				}

				// 检查是否在斜边附近（水平和垂直方向都有透明像素）
				hasHorizontalTransparent := mask.AlphaAt(px-1, py).A == 0 || mask.AlphaAt(px+1, py).A == 0
				hasVerticalTransparent := mask.AlphaAt(px, py-1).A == 0 || mask.AlphaAt(px, py+1).A == 0

				// 如果两个方向都有透明像素，可能是斜边
				if hasHorizontalTransparent && hasVerticalTransparent {
					// 收集更大范围的像素进行额外平滑
					var sumR, sumG, sumB uint32
					var totalWeight float64

					// 检查对角线方向和周围
					for dy := -2; dy <= 2; dy++ {
						for dx := -2; dx <= 2; dx++ {
							if dx == 0 && dy == 0 {
								continue
							}
							nx := px + dx
							ny := py + dy
							if nx >= 0 && nx < PuzzleWidth && ny >= 0 && ny < PuzzleHeight {
								if mask.AlphaAt(nx, ny).A > 0 {
									c := piece.RGBAAt(nx, ny)
									if !(c.R == 255 && c.G == 255 && c.B == 255) {
										distance := math.Sqrt(float64(dx*dx + dy*dy))
										weight := 1.0 / (distance + 1.0)

										sumR += uint32(float64(c.R) * weight)
										sumG += uint32(float64(c.G) * weight)
										sumB += uint32(float64(c.B) * weight)
										totalWeight += weight
									}
								}
							}
						}
					}

					if totalWeight > 0 {
						avgR := sumR / uint32(totalWeight)
						avgG := sumG / uint32(totalWeight)
						avgB := sumB / uint32(totalWeight)

						// 对斜边像素进行更强的平滑（60%混合）
						piece.SetRGBA(px, py, color.RGBA{
							R: uint8(float64(current.R)*0.4 + float64(avgR)*0.6),
							G: uint8(float64(current.G)*0.4 + float64(avgG)*0.6),
							B: uint8(float64(current.B)*0.4 + float64(avgB)*0.6),
							A: 255,
						})
					}
				}
			}
		}
	}
}

// countTransparentNeighbors 计算透明邻居数量
func countTransparentNeighbors(x, y int, mask *image.Alpha) int {
	count := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx := x + dx
			ny := y + dy
			if nx >= 0 && nx < PuzzleWidth && ny >= 0 && ny < PuzzleHeight {
				if mask.AlphaAt(nx, ny).A == 0 {
					count++
				}
			}
		}
	}
	return count
}

// isEdgeSimple 简单的边缘检测
func isEdgeSimple(x, y int, mask *image.Alpha) bool {
	// 检查周围3x3像素
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx := x + dx
			ny := y + dy
			if nx < 0 || nx >= PuzzleWidth || ny < 0 || ny >= PuzzleHeight {
				return true
			}
			if mask.AlphaAt(nx, ny).A == 0 {
				return true
			}
		}
	}
	return false
}

// add3DEffect 添加立体感效果（高光）
func add3DEffect(piece *image.RGBA, mask *image.Alpha) {
	// 对边缘内侧像素添加轻微的高光效果
	for py := 0; py < PuzzleHeight; py++ {
		for px := 0; px < PuzzleWidth; px++ {
			if mask.AlphaAt(px, py).A > 0 {
				// 检查是否在边缘
				transparentNeighbors := countTransparentNeighbors(px, py, mask)
				if transparentNeighbors > 0 {
					current := piece.RGBAAt(px, py)

					// 跳过白色边框
					if current.R == 255 && current.G == 255 && current.B == 255 {
						continue
					}

					// 根据透明邻居数量调整高光强度
					// 边缘越明显（透明邻居越多），高光越强
					highlightRatio := 0.05 + float64(transparentNeighbors)/9.0*0.15

					// 提高亮度，增加高光效果
					piece.SetRGBA(px, py, color.RGBA{
						R: clamp255(int(float64(current.R) * (1 + highlightRatio))),
						G: clamp255(int(float64(current.G) * (1 + highlightRatio))),
						B: clamp255(int(float64(current.B) * (1 + highlightRatio))),
						A: 255,
					})
				}
			}
		}
	}
}

// clamp255 限制值在0-255范围内
func clamp255(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// applyGaussianBlur 应用高斯模糊（多次应用增强版）
func applyGaussianBlur(piece *image.RGBA, mask *image.Alpha) {
	// 应用2次模糊，每次都基于上一次的结果
	for iteration := 0; iteration < 2; iteration++ {
		applyGaussianBlurOnce(piece, mask)
	}
}

// applyGaussianBlurToHole 对背景图上的缺口边缘应用高斯模糊
func applyGaussianBlurToHole(result *image.RGBA, mask *image.Alpha, offsetX, offsetY int) {
	// 创建副本用于模糊
	blurred := image.NewRGBA(result.Bounds())
	draw.Draw(blurred, result.Bounds(), result, image.Point{}, draw.Src)

	// 3x3 高斯核
	kernel := [3][3]float64{
		{1.0, 2.0, 1.0},
		{2.0, 4.0, 2.0},
		{1.0, 2.0, 1.0},
	}
	kernelSum := 16.0

	// 对缺口区域应用2次模糊
	for iteration := 0; iteration < 2; iteration++ {
		for py := 0; py < PuzzleHeight; py++ {
			for px := 0; px < PuzzleWidth; px++ {
				// 只处理mask内的像素
				if mask.AlphaAt(px, py).A > 0 {
					targetX := offsetX + px
					targetY := offsetY + py

					// 检查边界
					if targetX < 0 || targetX >= result.Bounds().Dx() ||
						targetY < 0 || targetY >= result.Bounds().Dy() {
						continue
					}

					var sumR, sumG, sumB float64

					// 应用3x3高斯核
					for ky := -1; ky <= 1; ky++ {
						for kx := -1; kx <= 1; kx++ {
							nx := targetX + kx
							ny := targetY + ky

							// 边界处理
							if nx < 0 {
								nx = 0
							}
							if nx >= result.Bounds().Dx() {
								nx = result.Bounds().Dx() - 1
							}
							if ny < 0 {
								ny = 0
							}
							if ny >= result.Bounds().Dy() {
								ny = result.Bounds().Dy() - 1
							}

							c := blurred.RGBAAt(nx, ny)
							weight := kernel[ky+1][kx+1]
							sumR += float64(c.R) * weight
							sumG += float64(c.G) * weight
							sumB += float64(c.B) * weight
						}
					}

					// 设置模糊后的像素
					result.SetRGBA(targetX, targetY, color.RGBA{
						R: uint8(sumR / kernelSum),
						G: uint8(sumG / kernelSum),
						B: uint8(sumB / kernelSum),
						A: 255,
					})
				}
			}
		}
		// 更新blurred为当前结果
		draw.Draw(blurred, result.Bounds(), result, image.Point{}, draw.Src)
	}
}

// applyGaussianBlurOnce 应用一次高斯模糊
func applyGaussianBlurOnce(piece *image.RGBA, mask *image.Alpha) {
	// 创建一个新的图像来存储模糊后的结果
	blurred := image.NewRGBA(piece.Bounds())

	// 3x3 高斯核
	kernel := [3][3]float64{
		{1.0, 2.0, 1.0},
		{2.0, 4.0, 2.0},
		{1.0, 2.0, 1.0},
	}
	kernelSum := 16.0

	for py := 0; py < PuzzleHeight; py++ {
		for px := 0; px < PuzzleWidth; px++ {
			// 只处理mask内的像素
			if mask.AlphaAt(px, py).A > 0 {
				var sumR, sumG, sumB float64

				// 应用3x3高斯核
				for ky := -1; ky <= 1; ky++ {
					for kx := -1; kx <= 1; kx++ {
						nx := px + kx
						ny := py + ky

						// 边界处理：使用边界像素
						if nx < 0 {
							nx = 0
						}
						if nx >= PuzzleWidth {
							nx = PuzzleWidth - 1
						}
						if ny < 0 {
							ny = 0
						}
						if ny >= PuzzleHeight {
							ny = PuzzleHeight - 1
						}

						// 只考虑mask内的像素
						if mask.AlphaAt(nx, ny).A > 0 {
							c := piece.RGBAAt(nx, ny)
							weight := kernel[ky+1][kx+1]
							sumR += float64(c.R) * weight
							sumG += float64(c.G) * weight
							sumB += float64(c.B) * weight
						}
					}
				}

				// 归一化并设置模糊后的像素
				blurred.SetRGBA(px, py, color.RGBA{
					R: uint8(sumR / kernelSum),
					G: uint8(sumG / kernelSum),
					B: uint8(sumB / kernelSum),
					A: 255,
				})
			} else {
				// 透明区域保持透明
				blurred.SetRGBA(px, py, color.RGBA{R: 0, G: 0, B: 0, A: 0})
			}
		}
	}

	// 将模糊后的图像复制回原图
	draw.Draw(piece, piece.Bounds(), blurred, image.Point{}, draw.Src)
}

/*
mask图片流程：
- 使用 64 倍于目标大小的高分辨率渲染
- 绘制多层半透明黑色阴影
- 对阴影进行高斯模糊
- 用双重形状（背景+前景）绘制以获得平滑边缘
- 双线性插值缩放到 70x70 大小
- 保存为透明背景的 PNG 文件
*/
