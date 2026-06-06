#ifndef AUTOGO_IMAGE_TEST_OPENCV_MATCH_BRIDGE_H
#define AUTOGO_IMAGE_TEST_OPENCV_MATCH_BRIDGE_H

#ifdef __cplusplus
extern "C" {
#endif

typedef struct OpenCVMatchPoint {
	int x;
	int y;
	float score;
} OpenCVMatchPoint;

typedef struct OpenCVMatchResult {
	int count;
	float best_score;
	int error_code;
	char error[256];
} OpenCVMatchResult;

OpenCVMatchResult autogo_match_template(
	const unsigned char* source_data,
	int source_width,
	int source_height,
	int source_stride,
	const unsigned char* template_data,
	int template_width,
	int template_height,
	int template_stride,
	int x1,
	int y1,
	int x2,
	int y2,
	int is_gray,
	int is_transparent,
	float sim,
	int find_all,
	OpenCVMatchPoint* points,
	int max_points
);

#ifdef __cplusplus
}
#endif

#endif
