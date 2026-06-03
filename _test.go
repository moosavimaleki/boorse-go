package main

func BenchmarkInBounds(b *testing.B) {
	for i := 0; i < b.N; i++ {
		main()
	}
}
