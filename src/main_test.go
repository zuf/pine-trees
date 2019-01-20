package main

//func BenchmarkDecodeThumb(b *testing.B) {
//	r := raw.NewRawProcessor()
//	defer r.Close()
//
//	for i := 0; i < b.N; i++ {
//		DecodeThumb(r, "./samples/TEST.CR2")
//	}
//}
//
//func BenchmarkDecodeThumbParallel(b *testing.B) {
//	b.RunParallel(func(pb *testing.PB) {
//		r := raw.NewRawProcessor()
//		defer r.Close()
//		for pb.Next() {
//			DecodeThumb(r, "./samples/TEST.CR2")
//		}
//	})
//}
