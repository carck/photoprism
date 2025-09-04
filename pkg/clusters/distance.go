package clusters

import "unsafe"

/*
#include <stdlib.h>
#include <stdio.h>
#include <math.h>
#include <arm_neon.h>

// left_ptr: 当前点 ai 的指针
// right_ptrs: [start:end] 向量指针数组
// end_len: right_ptrs 的长度
// res: 输出平方距离数组
void EuclideanDistance512(
    float *left_ptr, float **right_ptrs, float *res, int end_len, float eps_sq
) {
    for (int j = 0; j < end_len; j++) {
        float *right = right_ptrs[j];
        float dist_sq = 0.0f;

        for (int i = 0; i < 512; i += 8) {
            // load 8 floats at once
            float32x4x2_t a_vec = vld1q_f32_x2(left_ptr + i);
            float32x4x2_t b_vec = vld1q_f32_x2(right + i);

            float32x4_t diff0 = vsubq_f32(a_vec.val[0], b_vec.val[0]);
            float32x4_t diff1 = vsubq_f32(a_vec.val[1], b_vec.val[1]);

            float32x4_t sq0 = vmulq_f32(diff0, diff0);
            float32x4_t sq1 = vmulq_f32(diff1, diff1);

            // horizontal add both
            dist_sq += vaddvq_f32(sq0) + vaddvq_f32(sq1);

            if (dist_sq > eps_sq) {
                dist_sq = INFINITY;
                break;
            }
        }
        res[j] = dist_sq;
    }
}
*/
// #cgo CFLAGS: -O3 -ffast-math
// #cgo nocallback EuclideanDistance512
// #cgo noescape EuclideanDistance512
import "C"

func EuclideanDistance512C(d [][]float32, ai, start, end int, eps_sq float32) []float32 {
	length := end - start
	res := make([]float32, length)

	// 左向量 ai
	aiPtr := unsafe.Pointer(&d[ai][0])

	// 右向量 [start:end]
	rightPtrs := make([]uintptr, length)
	for i := start; i < end; i++ {
		rightPtrs[i-start] = uintptr(unsafe.Pointer(&d[i][0]))
	}

	C.EuclideanDistance512(
		(*C.float)(aiPtr),
		(**C.float)(unsafe.Pointer(&rightPtrs[0])),
		(*C.float)(unsafe.Pointer(&res[0])),
		C.int(length),
		C.float(eps_sq),
	)
	return res
}
