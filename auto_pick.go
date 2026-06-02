package main

import (
	"image"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	autoPickModeRandom            = "随机取点"
	autoPickModeContour           = "轮廓取点"
	autoPickModeHighlight         = "高亮取点"
	autoPickModeHighSaturation    = "高饱和取点"
	autoPickModeColorClassContour = "颜色分类轮廓"
	autoPickModeColorClassRandom  = "颜色分类随机"
	defaultAutoPickCount          = 20
)

type autoPickCandidate struct {
	Point image.Point
	Score float64
	Class int
}

type colorBucketStat struct {
	Class int
	Count int
}

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
	case autoPickModeContour:
		return pickContourPoints(req.Image, rect, req.Count, minDistance)
	case autoPickModeHighlight:
		return pickHighlightPoints(req.Image, rect, req.Count, minDistance)
	case autoPickModeHighSaturation:
		return pickHighSaturationPoints(req.Image, rect, req.Count, minDistance)
	case autoPickModeColorClassContour:
		return pickColorClassContourPoints(req.Image, rect, req.Count, minDistance)
	case autoPickModeColorClassRandom:
		return pickColorClassRandomPoints(req.Image, rect, req.Count, minDistance, req.Rand)
	default:
		return nil
	}
}

func supportedAutoPickMode(mode string) bool {
	switch mode {
	case autoPickModeRandom, autoPickModeContour, autoPickModeHighlight, autoPickModeHighSaturation, autoPickModeColorClassContour, autoPickModeColorClassRandom:
		return true
	default:
		return false
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

func pickContourPoints(img image.Image, rect image.Rectangle, count, minDistance int) []image.Point {
	if img == nil || rect.Empty() || count <= 0 {
		return nil
	}

	candidates := make([]autoPickCandidate, 0, rect.Dx()*rect.Dy())
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			left := pickLumaAt(img, max(rect.Min.X, x-1), y)
			right := pickLumaAt(img, min(rect.Max.X-1, x+1), y)
			top := pickLumaAt(img, x, max(rect.Min.Y, y-1))
			bottom := pickLumaAt(img, x, min(rect.Max.Y-1, y+1))
			score := math.Abs(right-left) + math.Abs(bottom-top)
			if score <= 0 {
				continue
			}

			candidates = append(candidates, autoPickCandidate{
				Point: image.Pt(x, y),
				Score: score,
			})
		}
	}

	return pickTopCandidates(candidates, count, minDistance)
}

func pickHighlightPoints(img image.Image, rect image.Rectangle, count, minDistance int) []image.Point {
	if img == nil || rect.Empty() || count <= 0 {
		return nil
	}

	candidates := make([]autoPickCandidate, 0, rect.Dx()*rect.Dy())
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			luma := pickLumaAt(img, x, y)
			if luma < 40 {
				continue
			}

			contrast := pickLocalLumaContrast(img, rect, x, y)
			score := luma*0.7 + contrast*0.3
			if score <= 0 {
				continue
			}

			candidates = append(candidates, autoPickCandidate{
				Point: image.Pt(x, y),
				Score: score,
			})
		}
	}

	return pickTopCandidates(candidates, count, minDistance)
}

func pickHighSaturationPoints(img image.Image, rect image.Rectangle, count, minDistance int) []image.Point {
	if img == nil || rect.Empty() || count <= 0 {
		return nil
	}

	candidates := make([]autoPickCandidate, 0, rect.Dx()*rect.Dy())
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			saturation, ok := pickSaturationAt(img, x, y)
			if !ok || saturation <= 0 {
				continue
			}

			contrast := pickLocalLumaContrast(img, rect, x, y)
			score := saturation*0.75 + contrast*0.25
			if score <= 0 {
				continue
			}

			candidates = append(candidates, autoPickCandidate{
				Point: image.Pt(x, y),
				Score: score,
			})
		}
	}

	return pickTopCandidates(candidates, count, minDistance)
}

func pickColorClassContourPoints(img image.Image, rect image.Rectangle, count, minDistance int) []image.Point {
	if img == nil || rect.Empty() || count <= 0 {
		return nil
	}

	stats := pickTopColorBuckets(img, rect, maxColorClassBuckets(count))
	if len(stats) == 0 {
		return nil
	}

	allowed := colorBucketSet(stats)
	countByClass := colorBucketCountMap(stats)
	candidatesByClass := make(map[int][]autoPickCandidate, len(stats))
	allCandidates := make([]autoPickCandidate, 0)
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			class, ok := pickColorBucketAt(img, x, y)
			if !ok || !allowed[class] || !colorBucketBoundaryAt(img, rect, x, y, class) {
				continue
			}

			left := pickLumaAt(img, max(rect.Min.X, x-1), y)
			right := pickLumaAt(img, min(rect.Max.X-1, x+1), y)
			top := pickLumaAt(img, x, max(rect.Min.Y, y-1))
			bottom := pickLumaAt(img, x, min(rect.Max.Y-1, y+1))
			score := math.Abs(right-left) + math.Abs(bottom-top) + math.Log1p(float64(countByClass[class]))
			candidate := autoPickCandidate{Point: image.Pt(x, y), Score: score, Class: class}
			candidatesByClass[class] = append(candidatesByClass[class], candidate)
			allCandidates = append(allCandidates, candidate)
		}
	}

	quotas := colorClassQuotas(stats, count)
	points := make([]image.Point, 0, count)
	seen := make(map[image.Point]struct{}, count)
	for _, stat := range stats {
		targetCount := len(points) + quotas[stat.Class]
		appendTopPickCandidates(&points, seen, candidatesByClass[stat.Class], targetCount, minDistance)
	}
	if len(points) < count {
		appendTopPickCandidates(&points, seen, allCandidates, count, minDistance)
	}
	return points
}

func pickColorClassRandomPoints(img image.Image, rect image.Rectangle, count, minDistance int, rng *rand.Rand) []image.Point {
	if img == nil || rect.Empty() || count <= 0 {
		return nil
	}
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	stats := pickTopColorBuckets(img, rect, maxColorClassBuckets(count))
	if len(stats) == 0 {
		return nil
	}

	allowed := colorBucketSet(stats)
	pointsByClass := make(map[int][]image.Point, len(stats))
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			class, ok := pickColorBucketAt(img, x, y)
			if ok && allowed[class] {
				pointsByClass[class] = append(pointsByClass[class], image.Pt(x, y))
			}
		}
	}

	quotas := colorClassQuotas(stats, count)
	points := make([]image.Point, 0, count)
	seen := make(map[image.Point]struct{}, count)
	for _, stat := range stats {
		appendRandomPickPoints(&points, seen, pointsByClass[stat.Class], quotas[stat.Class], minDistance, rng)
	}
	if len(points) < count {
		for _, stat := range stats {
			appendScanPickPoints(&points, seen, pointsByClass[stat.Class], count, minDistance)
			if len(points) >= count {
				break
			}
		}
	}
	return points
}

func pickTopCandidates(candidates []autoPickCandidate, count, minDistance int) []image.Point {
	if len(candidates) == 0 || count <= 0 {
		return nil
	}

	sortAutoPickCandidates(candidates)

	points := make([]image.Point, 0, min(count, len(candidates)))
	seen := make(map[image.Point]struct{}, count)
	appendTopPickCandidates(&points, seen, candidates, count, minDistance)
	return points
}

func sortAutoPickCandidates(candidates []autoPickCandidate) {
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		if candidates[i].Point.Y != candidates[j].Point.Y {
			return candidates[i].Point.Y < candidates[j].Point.Y
		}
		return candidates[i].Point.X < candidates[j].Point.X
	})
}

func appendTopPickCandidates(points *[]image.Point, seen map[image.Point]struct{}, candidates []autoPickCandidate, targetCount, minDistance int) {
	sortAutoPickCandidates(candidates)
	for _, candidate := range candidates {
		if len(*points) >= targetCount {
			break
		}
		if _, exists := seen[candidate.Point]; exists || !pointDistanceOK(*points, candidate.Point, minDistance) {
			continue
		}
		seen[candidate.Point] = struct{}{}
		*points = append(*points, candidate.Point)
	}
}

func pickLumaAt(img image.Image, x, y int) float64 {
	r, g, b, _ := img.At(x, y).RGBA()
	return 0.299*float64(uint8(r>>8)) + 0.587*float64(uint8(g>>8)) + 0.114*float64(uint8(b>>8))
}

func pickSaturationAt(img image.Image, x, y int) (float64, bool) {
	r, g, b, a := img.At(x, y).RGBA()
	if uint8(a>>8) < 32 {
		return 0, false
	}

	r8 := uint8(r >> 8)
	g8 := uint8(g >> 8)
	b8 := uint8(b >> 8)
	maxValue := max(max(int(r8), int(g8)), int(b8))
	if maxValue == 0 {
		return 0, true
	}

	minValue := min(min(int(r8), int(g8)), int(b8))
	return float64(maxValue-minValue) / float64(maxValue) * 255, true
}

func pickLocalLumaContrast(img image.Image, rect image.Rectangle, x, y int) float64 {
	center := pickLumaAt(img, x, y)
	total := 0.0
	count := 0
	addNeighbor := func(nx, ny int) {
		if nx < rect.Min.X || nx >= rect.Max.X || ny < rect.Min.Y || ny >= rect.Max.Y {
			return
		}
		total += pickLumaAt(img, nx, ny)
		count++
	}

	addNeighbor(x-1, y)
	addNeighbor(x+1, y)
	addNeighbor(x, y-1)
	addNeighbor(x, y+1)
	if count == 0 {
		return 0
	}
	return math.Abs(center - total/float64(count))
}

func maxColorClassBuckets(count int) int {
	if count <= 0 {
		return 0
	}
	if count < 6 {
		return count
	}
	return 6
}

func pickTopColorBuckets(img image.Image, rect image.Rectangle, maxBuckets int) []colorBucketStat {
	if img == nil || rect.Empty() || maxBuckets <= 0 {
		return nil
	}

	counts := make(map[int]int)
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			class, ok := pickColorBucketAt(img, x, y)
			if ok {
				counts[class]++
			}
		}
	}

	stats := make([]colorBucketStat, 0, len(counts))
	for class, count := range counts {
		stats = append(stats, colorBucketStat{Class: class, Count: count})
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count != stats[j].Count {
			return stats[i].Count > stats[j].Count
		}
		return stats[i].Class < stats[j].Class
	})
	if len(stats) > maxBuckets {
		stats = stats[:maxBuckets]
	}
	return stats
}

func pickColorBucketAt(img image.Image, x, y int) (int, bool) {
	r, g, b, a := img.At(x, y).RGBA()
	if uint8(a>>8) < 32 {
		return 0, false
	}

	rBucket := int(uint8(r>>8)) / 32
	gBucket := int(uint8(g>>8)) / 32
	bBucket := int(uint8(b>>8)) / 32
	return rBucket<<6 | gBucket<<3 | bBucket, true
}

func colorBucketSet(stats []colorBucketStat) map[int]bool {
	set := make(map[int]bool, len(stats))
	for _, stat := range stats {
		set[stat.Class] = true
	}
	return set
}

func colorBucketCountMap(stats []colorBucketStat) map[int]int {
	counts := make(map[int]int, len(stats))
	for _, stat := range stats {
		counts[stat.Class] = stat.Count
	}
	return counts
}

func colorClassQuotas(stats []colorBucketStat, total int) map[int]int {
	quotas := make(map[int]int, len(stats))
	if len(stats) == 0 || total <= 0 {
		return quotas
	}

	active := stats
	if len(active) > total {
		active = active[:total]
	}
	for _, stat := range active {
		quotas[stat.Class] = 1
	}

	remaining := total - len(active)
	for remaining > 0 {
		for _, stat := range active {
			if remaining == 0 {
				break
			}
			quotas[stat.Class]++
			remaining--
		}
	}
	return quotas
}

func colorBucketBoundaryAt(img image.Image, rect image.Rectangle, x, y, class int) bool {
	neighbors := [][2]int{
		{x - 1, y},
		{x + 1, y},
		{x, y - 1},
		{x, y + 1},
	}
	for _, neighbor := range neighbors {
		nx, ny := neighbor[0], neighbor[1]
		if nx < rect.Min.X || nx >= rect.Max.X || ny < rect.Min.Y || ny >= rect.Max.Y {
			continue
		}
		neighborClass, ok := pickColorBucketAt(img, nx, ny)
		if !ok || neighborClass != class {
			return true
		}
	}
	return false
}

func appendRandomPickPoints(points *[]image.Point, seen map[image.Point]struct{}, pool []image.Point, targetAdd, minDistance int, rng *rand.Rand) {
	if targetAdd <= 0 || len(pool) == 0 {
		return
	}

	targetCount := len(*points) + targetAdd
	maxAttempts := max(targetAdd*80, len(pool)*2)
	for attempts := 0; attempts < maxAttempts && len(*points) < targetCount; attempts++ {
		p := pool[rng.Intn(len(pool))]
		if _, exists := seen[p]; exists || !pointDistanceOK(*points, p, minDistance) {
			continue
		}
		seen[p] = struct{}{}
		*points = append(*points, p)
	}
	if len(*points) < targetCount {
		appendScanPickPoints(points, seen, pool, targetCount, minDistance)
	}
}

func appendScanPickPoints(points *[]image.Point, seen map[image.Point]struct{}, pool []image.Point, targetCount, minDistance int) {
	for _, p := range pool {
		if len(*points) >= targetCount {
			return
		}
		if _, exists := seen[p]; exists || !pointDistanceOK(*points, p, minDistance) {
			continue
		}
		seen[p] = struct{}{}
		*points = append(*points, p)
	}
}
