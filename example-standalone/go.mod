module example-standalone

go 1.21

require github.com/yourusername/go-analysis/client v0.0.0

require github.com/google/uuid v1.6.0 // indirect

replace github.com/yourusername/go-analysis/client => ../
