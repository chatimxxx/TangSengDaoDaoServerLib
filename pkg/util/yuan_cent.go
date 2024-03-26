package util

import "fmt"

// YuanToCent 元转分
func YuanToCent(yuan float64) (int64, error) {
	dec, err := NewFromString(fmt.Sprintf("%0.2f", yuan))
	if err != nil {
		return 0, err
	}
	m, err := NewFromString("100")
	if err != nil {
		return 0, err
	}
	return dec.Mul(m).IntPart(), nil
}

// CentToYuan 分转元
func CentToYuan(cent int64) (float64, error) {
	centDec, err := NewFromString(fmt.Sprintf("%d", cent))
	if err != nil {
		return 0, err
	}
	mDec, err := NewFromString(fmt.Sprintf("%d", 100))
	if err != nil {
		return 0, err
	}
	result, ok := centDec.Div(mDec).Round(2).Float64()
	if ok {
		return 0, err
	}
	return result, nil
}
