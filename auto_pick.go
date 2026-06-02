package main

import (
	"image"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	autoPickModeRandom   = "随机取点"
	defaultAutoPickCount = 20
)

type autoPickRequest struct {
	Image       image.Image
	Rect        image.Rectangle
	Count       int
	Mode        string
	MinDistance int
	Rand        *rand.Rand
}

func autoPickPoints(req autoPickRequest) []image.Point {
	rect := normalizePickRect(req.Image, req.Rect)
	if rect.Empty() || req.Count <= 0 {
		return nil
	}

	minDistance := req.MinDistance
	if minDistance <= 0 {
		minDistance = estimatePickMinDistance(rect, req.Count)
	}

	switch req.Mode {
	case autoPickModeRandom:
		return pickRandomPoints(rect, req.Count, minDistance, req.Rand)
	default:
		return nil
	}
}

func parsePickCount(text string) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return defaultAutoPickCount
	}

	var digits strings.Builder
	for _, r := range text {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
			continue
		}
		if digits.Len() > 0 {
			break
		}
	}
	if digits.Len() == 0 {
		return defaultAutoPickCount
	}

	count, err := strconv.Atoi(digits.String())
	if err != nil {
		return defaultAutoPickCount
	}
	return count
}

func normalizePickRect(img image.Image, rect image.Rectangle) image.Rectangle {
	if img == nil {
		return image.Rectangle{}
	}

	minX := min(rect.Min.X, rect.Max.X)
	minY := min(rect.Min.Y, rect.Max.Y)
	maxX := max(rect.Min.X, rect.Max.X)
	maxY := max(rect.Min.Y, rect.Max.Y)
	normalized := image.Rect(minX, minY, maxX, maxY).Intersect(img.Bounds())
	if normalized.Dx() <= 0 || normalized.Dy() <= 0 {
		return image.Rectangle{}
	}
	return normalized
}

func inclusivePickRect(x1, y1, x2, y2 int) image.Rectangle {
	minX := min(x1, x2)
	minY := min(y1, y2)
	maxX := max(x1, x2)
	maxY := max(y1, y2)
	return image.Rect(minX, minY, maxX+1, maxY+1)
}

func pointDistanceOK(points []image.Point, p image.Point, minDistance int) bool {
	if minDistance <= 0 {
		return true
	}

	minDistanceSquared := minDistance * minDistance
	for _, point := range points {
		dx := p.X - point.X
		dy := p.Y - point.Y
		if dx*dx+dy*dy < minDistanceSquared {
			return false
		}
	}
	return true
}

func estimatePickMinDistance(rect image.Rectangle, count int) int {
	if count <= 0 || rect.Empty() {
		return 3
	}

	area := rect.Dx() * rect.Dy()
	if area <= 0 {
		return 3
	}

	distance := int(math.Sqrt(float64(area)/float64(count)) / 2)
	if distance < 3 {
		return 3
	}
	return distance
}

func pickRandomPoints(rect image.Rectangle, count, minDistance int, rng *rand.Rand) []image.Point {
	if rect.Empty() || count <= 0 {
		return nil
	}
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	maxCount := rect.Dx() * rect.Dy()
	if count > maxCount {
		count = maxCount
	}

	points := make([]image.Point, 0, count)
	seen := make(map[image.Point]struct{}, count)
	maxAttempts := count * 80
	for attempts := 0; attempts < maxAttempts && len(points) < count; attempts++ {
		p := image.Pt(
			rect.Min.X+rng.Intn(rect.Dx()),
			rect.Min.Y+rng.Intn(rect.Dy()),
		)
		if !appendPickPoint(&points, seen, p, minDistance) {
			continue
		}
	}

	if len(points) < count {
		gridScanPickPoints(rect, count, minDistance, seen, &points)
	}
	return points
}

func gridScanPickPoints(rect image.Rectangle, count, minDistance int, seen map[image.Point]struct{}, points *[]image.Point) {
	if len(*points) >= count {
		return
	}

	step := int(math.Sqrt(float64(rect.Dx()*rect.Dy()) / float64(count)))
	if step < 1 {
		step = 1
	}

	for y := rect.Min.Y; y < rect.Max.Y && len(*points) < count; y += step {
		for x := rect.Min.X; x < rect.Max.X && len(*points) < count; x += step {
			appendPickPoint(points, seen, image.Pt(x, y), minDistance)
		}
	}
	for y := rect.Min.Y; y < rect.Max.Y && len(*points) < count; y++ {
		for x := rect.Min.X; x < rect.Max.X && len(*points) < count; x++ {
			appendPickPoint(points, seen, image.Pt(x, y), minDistance)
		}
	}
}

func appendPickPoint(points *[]image.Point, seen map[image.Point]struct{}, p image.Point, minDistance int) bool {
	if _, exists := seen[p]; exists {
		return false
	}
	if !pointDistanceOK(*points, p, minDistance) {
		return false
	}

	seen[p] = struct{}{}
	*points = append(*points, p)
	return true
}
