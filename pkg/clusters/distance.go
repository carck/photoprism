package clusters

import "unsafe"

/*
#include <stdlib.h>
#include <stdio.h>
#include <math.h>
//#include <arm_neon.h>

void EuclideanDistance512(float **d, float *res, int ai, int bi, int end) {
        float s,t;
        float *left = d[ai];
        int c = 0;
        for (int j = bi; j < end; j++ ){
                s = 0;
                float *right = d[j];
                //float32x4_t sum = vdupq_n_f32(0);
                for (int i = 0; i < 512; i++) {
                        //float32x4_t il = vld1q_f32(left+i);
                        //float32x4_t ir = vld1q_f32(right+i);
                        //float32x4_t sub = vsubq_f32(il, ir);
                        //sum = vfmaq_f32(sum, sub, sub);
                        t = left[i] - right[i];
                        s += t * t;
                }
                //s = vaddvq_f32(sum);
                res[c++] = (float)sqrt(s);
        }
}
*/
// #cgo CFLAGS: -O3 -ffast-math
import "C"

func EuclideanDistance512C(d [][]float32, ai, bi, end int) []float32 {
	res := make([]float32, end-bi)

	data := make([]uintptr, len(d))
	keep := make([]unsafe.Pointer, len(d))

	for i, v := range d {
		p := unsafe.Pointer(&v[0])
		keep[i] = p
		data[i] = uintptr(p)
	}

	C.EuclideanDistance512(
		(**C.float)(unsafe.Pointer(&data[0])),
		(*C.float)(unsafe.Pointer(&res[0])),
		C.int(ai),
		C.int(bi),
		C.int(end),
	)
	return res
}
