package io

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"

type MockedFeeder struct {
	pts []*point.Point

	lastErrors [][2]string
}

func NewMockedFeeder() *MockedFeeder {
	return &MockedFeeder{}
}

func (f *MockedFeeder) Feed(name, category string, pts []*point.Point, opt ...*Option) error {
	// TODO: run pipeline & filter

	f.pts = append(f.pts, pts...)
	return nil
}

func (f *MockedFeeder) FeedLastError(name, errInfo string) {
	f.lastErrors = append(f.lastErrors, [2]string{name, errInfo})
}

func (f *MockedFeeder) Clear() {
	f.pts = f.pts[:0]
	f.lastErrors = f.lastErrors[:0]
}

func (f *MockedFeeder) Points() []*point.Point {
	return f.pts
}

func (f *MockedFeeder) LastErrors() [][2]string {
	return f.lastErrors
}
