package utils

import "sync/atomic"

func GetAndAddInt32(target *int32) int32 {
	return GetAndAddInt32WithDelta(target, 1)
}

func GetAndAddInt32WithDelta(target *int32, delta int32) int32 {
	for {
		expect := *target
		newVal := expect + delta
		swapSuccess := atomic.CompareAndSwapInt32(target, expect, newVal)
		if swapSuccess {
			return expect
		}
	}
}

func GetCyclic(target *int32, delta int32, max int32, base int32) int32 {
	result := GetAndAddInt32WithDelta(target, delta)

	if max > result {
		return result
	}

	for {
		nowVal := *target
		if max > nowVal {
			return nowVal
		}

		swapSuccess := atomic.CompareAndSwapInt32(target, nowVal, base+delta)
		if swapSuccess {
			return base
		}
	}
}