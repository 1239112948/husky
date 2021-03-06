package util

import (
	"bytes"
	"encoding/json"
	"testing"
)

type AA struct {
	N1 int64
	N2 int32
	S1 string
	F1 float32
	B1 bool
	B2 bool
	M1 map[string]string
	M2 map[string]string
	A1 []int
}

type AB struct {
	N1 int32
	N2 int64
	S1 string
	F1 float64
	B1 bool
	M1 map[string]string
	M2 map[string]string
	n1 int32
	N3 int
	A1 []int
}

type A struct {
	N1  int64
	N2  int32
	S1  string
	F1  float32
	B1  bool
	M3  map[string]string
	M4  map[string]string
	AA1 AA
	AA3 *AA
	aa1 AA
	AA4 []AA
	AA5 []*AA
	AA6 []int
}

type B struct {
	N1  int64
	N3  int64
	S1  string
	F1  float32
	F2  float64
	M4  map[string]string
	AA1 *AB
	AA2 AB
	AA3 AA
	aa1 *AA
	AA4 []AA
	AA5 []AA
	AA6 []int
}

func TestSructCopy(t *testing.T) {
	aa3 := AA{N1: 31, N2: 32, S1: "AAS3", B1: true}
	aa4 := AA{N1: 41, N2: 42, S1: "AAS4", B1: false}
	a := &A{
		N1: 1, N2: 2,
		S1:  "S1",
		AA1: AA{N1: 11, N2: 12, S1: "AAS1", B1: true},
		AA3: &AA{N1: 21, N2: 22, S1: "AAS2", B1: false},
		AA4: []AA{aa3, aa4},
		AA5: []*AA{&aa3, &aa4},
		AA6: []int{1, 2, 3},
	}
	b1 := &B{}
	b2 := &B{}
	DeepCopy(b1, a)
	s1, _ := json.Marshal(b1)
	s2, _ := json.Marshal(a)
	json.Unmarshal(s2, b2)
	s2, _ = json.Marshal(b2)
	t.Log(string(s1))
	t.Log(string(s2))
	DeepCopy(nil, a)
	DeepCopy(nil, nil)
	if bytes.Compare(s1, s2) != 0 {
		t.Error("deep copy error", string(s1), string(s2))
	}
}
