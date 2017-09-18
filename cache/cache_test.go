// cache测试
//   变更历史
//     2017-04-26  lixiaoya  新建
package cache

import (
	"testing"
	"fmt"
	"encoding/json"
)

// TestEncode Encode与Decode测试
func TestEncode(t *testing.T) {
	src := "HelloWorld!"
	key := "123456"
	enc, err := Encode([]byte(src), []byte(key))
	if err != nil {
		t.Errorf("Encode failed. err: %s.", err.Error())
		return
	}

	dec, err := Decode(enc, []byte(key))
	if err != nil {
		t.Errorf("Decode failed. err: %s.", err.Error())
		return
	} else if src != string(dec) {
		t.Errorf("Redis Get failed. Got %s, expected %s.", string(dec), src)
		return
	}

}

// TestTo InterToByte和ByteToInter测试
func TestTo(t *testing.T) {
	fmt.Println("string start >>>")
	sSrc1 := "HelloWorld!"
	b, err := InterToByte(sSrc1)
	if err != nil {
		t.Errorf("Encode InterToByte. err: %s.", err.Error())
		return
	}

	var sDst1 string
	err = ByteToInter(b, &sDst1)
	if err != nil {
		t.Errorf("Decode failed. err: %s.", err.Error())
		return
	} else if sSrc1 != sDst1 {
		t.Errorf("Redis Get failed. Got %s, expected %s.", sDst1, sSrc1)
		return
	}

	////////////////////////////////////////////////
	fmt.Println("string point start >>>")
	sSrc2 := "HelloWorld!"
	b, err = InterToByte(&sSrc2)
	if err != nil {
		t.Errorf("Encode InterToByte. err: %s.", err.Error())
		return
	}

	var sDst2 string
	err = ByteToInter(b, &sDst2)
	if err != nil {
		t.Errorf("Decode failed. err: %s.", err.Error())
		return
	} else if sSrc2 != sDst2 {
		t.Errorf("Redis Get failed. Got %s, expected %s.", sDst2, sSrc2)
		return
	}

	////////////////////////////////////////////////
	fmt.Println("int start >>>")
	iSrc := 100
	b, err = InterToByte(iSrc)
	if err != nil {
		t.Errorf("Encode InterToByte. err: %s.", err.Error())
		return
	}

	var iDst int
	err = ByteToInter(b, &iDst)
	if err != nil {
		t.Errorf("Decode failed. err: %s.", err.Error())
		return
	} else if iSrc != iDst {
		t.Errorf("Redis Get failed. Got %d, expected %d.", iDst, iSrc)
		return
	}

	////////////////////////////////////////////////
	fmt.Println("float64 start >>>")
	fSrc := 10.5
	b, err = InterToByte(fSrc)
	if err != nil {
		t.Errorf("Encode InterToByte. err: %s.", err.Error())
		return
	}

	var fDst float64
	err = ByteToInter(b, &fDst)
	if err != nil {
		t.Errorf("Decode failed. err: %s.", err.Error())
		return
	} else if fSrc != fDst {
		t.Errorf("Redis Get failed. Got %f, expected %f.", fDst, fSrc)
		return
	}

	////////////////////////////////////////////////
	fmt.Println("struct1 start >>>")
	type user1 struct {
		Uid   int32
		Name  string
	}
	stSrc := user1{
		Uid:  100,
		Name: "Diego",
	}
	b, err = InterToByte(stSrc)
	if err != nil {
		t.Errorf("Encode InterToByte. err: %s.", err.Error())
		return
	}

	stDst := user1{}
	err = ByteToInter(b, &stDst)
	if err != nil {
		t.Errorf("Decode failed. err: %s.", err.Error())
		return
	} else if stSrc.Uid != stDst.Uid || stSrc.Name != stDst.Name {
		t.Errorf("Redis Get failed. Got %d-%s, expected %d-%s.", stDst.Uid, stDst.Name, stSrc.Uid,stSrc.Name)
		return
	}

	////////////////////////////////////////////////
	fmt.Println("struct2 start >>>")
	stSrc2 := user2{
		Uid:  100,
		Name: "Diego",
	}
	b, err = InterToByte(&stSrc2)
	if err != nil {
		t.Errorf("Encode InterToByte. err: %s.", err.Error())
		return
	}

	stDst2 := user2{}
	err = ByteToInter(b, &stDst2)
	if err != nil {
		t.Errorf("Decode failed. err: %s.", err.Error())
		return
	} else if stSrc2.Uid != stDst2.Uid || stSrc2.Name != stDst2.Name {
		t.Errorf("Redis Get failed. Got %d-%s, expected %d-%s.", stDst2.Uid, stDst2.Name, stSrc2.Uid,stSrc2.Name)
		return
	}
}

type user2 struct {
	Uid   int32
	Name  string
}

func (this *user2)MarshalJSON() ([]byte, error) {
	fmt.Println("user2 MarshalJSON")

	str := fmt.Sprintf(`{"uid":%d, "name":"%s"}`, this.Uid, this.Name)
	return []byte(str), nil
}

func (this *user2)UnmarshalJSON(data []byte) error {
	fmt.Println("user2 UnmarshalJSON")

	val := make(map[string]interface{})
	json.Unmarshal(data, &val)
	uid, _ := val["uid"]
	this.Uid = int32(uid.(float64))
	name, _ := val["name"]
	this.Name = name.(string)
	return nil
}
