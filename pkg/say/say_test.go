package say

import (
	"slices"
	"testing"
)

func TestConvertSameFormat(t *testing.T) {
	in := []int16{1, 2, 3, 4}
	if got := Convert(in, 22050, 1, 22050, 1); !slices.Equal(got, in) {
		t.Errorf("got %v, want %v", got, in)
	}
}

func TestConvertStereoToMono(t *testing.T) {
	got := Convert([]int16{100, 300, -100, -300}, 22050, 2, 22050, 1)
	if want := []int16{200, -200}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConvertMonoToStereo(t *testing.T) {
	got := Convert([]int16{5, 7}, 22050, 1, 22050, 2)
	if want := []int16{5, 5, 7, 7}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConvertDownsample(t *testing.T) {
	got := Convert([]int16{0, 10, 20, 30}, 44100, 1, 22050, 1)
	if want := []int16{0, 20}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConvertUpsample(t *testing.T) {
	got := Convert([]int16{0, 10}, 22050, 1, 44100, 1)
	if want := []int16{0, 5, 10, 10}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
