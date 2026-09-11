package service

import "testing"

func TestNormalizeImageCreatorRequestControls(t *testing.T) {
	t.Parallel()

	if got := normalizeImageCreatorRequestedOutputFormat("png"); got != "png" {
		t.Fatalf("png format = %q, want png", got)
	}
	if got := normalizeImageCreatorRequestedOutputFormat("jpeg"); got != "jpeg" {
		t.Fatalf("jpeg format = %q, want jpeg", got)
	}
	if got := normalizeImageCreatorRequestedOutputFormat("webp"); got != "webp" {
		t.Fatalf("webp format = %q, want webp", got)
	}
	if got := normalizeImageCreatorQuality("standard"); got != "medium" {
		t.Fatalf("standard quality = %q, want medium", got)
	}
	if got := normalizeImageCreatorQuality("high"); got != "high" {
		t.Fatalf("high quality = %q, want high", got)
	}
}

func TestValidateImageCreatorSize(t *testing.T) {
	t.Parallel()

	for _, size := range []string{"", "auto", "1024x1024", "1536x1024", "1024x1536", "3072x1024"} {
		if err := validateImageCreatorSize(size); err != nil {
			t.Fatalf("size %q rejected: %v", size, err)
		}
	}
	for _, size := range []string{"1024x341", "4096x1024", "1024x1024x1"} {
		if err := validateImageCreatorSize(size); err == nil {
			t.Fatalf("size %q accepted", size)
		}
	}
}

func TestImageCreatorSizeForResolutionAndAspectRatio(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, resolution, ratio, want string }{
		{"1K square", "1K", "1:1", "1024x1024"},
		{"1K three two", "1K", "3:2", "1536x1024"},
		{"1K two three", "1K", "2:3", "1024x1536"},
		{"2K square", "2K", "1:1", "2048x2048"},
		{"4K square", "4K", "1:1", "2880x2880"},
		{"4K landscape", "4K", "16:9", "3840x2160"},
		{"4K portrait", "4K", "9:16", "2160x3840"},
		{"2K three two", "2K", "3:2", "2048x1360"},
	}
	for _, tt := range tests {
		got := imageCreatorSizeForResolutionAndAspectRatio(tt.resolution, tt.ratio)
		if got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
		if err := validateImageCreatorSize(got); err != nil {
			t.Errorf("%s produced invalid size %q: %v", tt.name, got, err)
		}
	}
}
