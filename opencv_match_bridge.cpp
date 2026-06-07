//go:build opencv_cgo
// +build opencv_cgo

#include "opencv_match_bridge.h"

#include <cmath>
#include <cstring>
#include <opencv2/core.hpp>
#include <opencv2/imgproc.hpp>

namespace {

OpenCVMatchResult make_result() {
	OpenCVMatchResult result;
	result.count = 0;
	result.best_score = -1.0f;
	result.error_code = 0;
	result.error[0] = '\0';
	return result;
}

OpenCVMatchResult make_error(const char* message) {
	OpenCVMatchResult result = make_result();
	result.error_code = 1;
	std::strncpy(result.error, message, sizeof(result.error) - 1);
	result.error[sizeof(result.error) - 1] = '\0';
	return result;
}

void append_point(OpenCVMatchPoint* points, int max_points, OpenCVMatchResult* result, int x, int y, float score) {
	if (result->count >= max_points) {
		return;
	}
	points[result->count].x = x;
	points[result->count].y = y;
	points[result->count].score = score;
	result->count++;
}

} // namespace

extern "C" OpenCVMatchResult autogo_match_template(
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
) {
	if (source_data == nullptr || template_data == nullptr || points == nullptr) {
		return make_error("OpenCV 输入数据为空");
	}
	if (source_width <= 0 || source_height <= 0 || template_width <= 0 || template_height <= 0) {
		return make_error("OpenCV 输入尺寸无效");
	}
	if (x1 < 0 || y1 < 0 || x2 > source_width || y2 > source_height || x1 >= x2 || y1 >= y2) {
		return make_error("OpenCV 查找范围无效");
	}
	if (template_width > x2 - x1 || template_height > y2 - y1) {
		return make_error("OpenCV 模板大于查找范围");
	}
	if (max_points <= 0) {
		return make_error("OpenCV 结果容量无效");
	}

	try {
		cv::Mat source(source_height, source_width, CV_8UC4, const_cast<unsigned char*>(source_data), source_stride);
		cv::Mat templ(template_height, template_width, CV_8UC4, const_cast<unsigned char*>(template_data), template_stride);
		cv::Mat roi = source(cv::Rect(x1, y1, x2 - x1, y2 - y1));

		cv::Mat source_match = roi;
		cv::Mat template_match = templ;
		cv::Mat source_gray;
		cv::Mat template_gray;
		if (is_gray) {
			// AutoGo 官方封装对 4 通道 Mat 使用 BGRToGray；这里保持同样的通道解释。
			cv::cvtColor(roi, source_gray, cv::COLOR_BGR2GRAY);
			cv::cvtColor(templ, template_gray, cv::COLOR_BGR2GRAY);
			source_match = source_gray;
			template_match = template_gray;
		}

		cv::Mat mask;
		if (is_transparent) {
			mask = cv::Mat(template_height, template_width, CV_8UC1, cv::Scalar(255));
			const cv::Vec4b transparent = templ.at<cv::Vec4b>(0, 0);
			for (int y = 0; y < template_height; y++) {
				for (int x = 0; x < template_width; x++) {
					const cv::Vec4b pixel = templ.at<cv::Vec4b>(y, x);
					if (pixel[0] == transparent[0] && pixel[1] == transparent[1] && pixel[2] == transparent[2]) {
						mask.at<unsigned char>(y, x) = 0;
					}
				}
			}
			if (cv::countNonZero(mask) == 0) {
				return make_error("OpenCV 模板透明遮罩为空");
			}
		}

		cv::Mat match_result;
		if (mask.empty()) {
			cv::matchTemplate(source_match, template_match, match_result, cv::TM_CCOEFF_NORMED);
		} else {
			cv::matchTemplate(source_match, template_match, match_result, cv::TM_CCOEFF_NORMED, mask);
		}
		if (match_result.empty()) {
			return make_error("OpenCV 匹配结果为空");
		}
		cv::patchNaNs(match_result, -1.0);

		double min_val = 0;
		double max_val = 0;
		cv::Point min_loc;
		cv::Point max_loc;
		cv::minMaxLoc(match_result, &min_val, &max_val, &min_loc, &max_loc);

		OpenCVMatchResult result = make_result();
		result.best_score = static_cast<float>(max_val);

		const float threshold = sim;
		if (find_all) {
			for (int y = 0; y < match_result.rows; y++) {
				const float* row = match_result.ptr<float>(y);
				for (int x = 0; x < match_result.cols; x++) {
					const float score = row[x];
					if (std::isfinite(score) && score >= threshold) {
						append_point(points, max_points, &result, x1 + x, y1 + y, score);
						if (result.count >= max_points) {
							return result;
						}
					}
				}
			}
			return result;
		}

		if (std::isfinite(max_val) && max_val >= threshold) {
			append_point(points, max_points, &result, x1 + max_loc.x, y1 + max_loc.y, static_cast<float>(max_val));
		}
		return result;
	} catch (const cv::Exception& e) {
		return make_error(e.what());
	} catch (...) {
		return make_error("OpenCV 匹配发生未知错误");
	}
}
