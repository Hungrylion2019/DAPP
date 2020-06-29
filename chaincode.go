package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"math/rand"
	"strconv"
	"time"
)


const Admin = "Admin"
const Pwd = "123456"
const DeviceID  = "Zhuoer"

//定义智能合约结构体
type SmartContract struct {
}

const AdminName = "skyhuihui"
const TokenKey = "Token"

type TransactionRecord struct {
	From        string  `json:"From"`
	To          string  `json:"To"`
	TokenSymbol string  `json:"TokenSymbol"`
	Amount      float64 `json:"Amount"`
	TxId        string  `json:"TxId"`
}
type Currency struct {
	TokenName   string  `json:"TokenName"`
	TokenSymbol string  `json:"TokenSymbol"`
	TotalSupply float64 `json:"TotalSupply"`
	//User        map[string]float64  `json:"User"` //某代币下各个用户持有的数量   //删除，不需要
	Record []TransactionRecord `json:"Record"`
}
type Token struct {
	Currency map[string]Currency `json:"Currency"`
}

//A3用户交易产生的手续费 创建A3的A2用户可以获得分成  创建A2用户的A1用户也可以获得分成。B类相同
type Member struct {
	MemberID      int                `json:"member_id"`    //会员编号              //同一个类别存在多个用户，会员。
	MemberName    string             `json:"member_name"`  //会员姓名
	MemberPwd     string             `"json:member_pwd"`   //会员密码
	MemberClass   string             `json:"member_class"` //会员类别：ABC三类
	MemberLevel   int                `json:"member_level"` //会员级别
	DeviceID      string             `json:"device_id"`    //登录绑定设备号
	SafeCode      string             `"json:safe_code"`    //安全码
	BalanceOf     map[string]float64 `json:"BalanceOf"`    //对应币的数量
	Frozen        bool               `json:"Frozen"`       //账户是否冻结
	Charge        map[string]float64 `json:"Charge"`       //用户从下级某个币获得奖励分成
	SuperiorCount map[string]float64 `json:"Service"`      //能够享有分成的上级账户，同时和分成比例挂钩。
	Fee           float64            `json:"Fee"`          //能够享有分成的上级账户，同时和分成比例挂钩。//本身账户交易产生的比例  FEE
}

//如果需要比例分离，需要加入这个结构体
//type Fee struct {
//	A1fee float64 `json:"A1fee"`
//	A2fee float64 `json:"A2fee"`
//	A3fee float64 `json:"A3fee"`
//	B1fee float64 `json:"B1fee"`
//	B2fee float64 `json:"B2fee"`
//	B3fee float64 `json:"B3fee"`
//	B4fee float64 `json:"B4fee"`
//	Cfee  float64 `json:"Cfee"`
//}

//Invoke函数
func (t *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	//peer chaincode invoke -C mychannel -n mycc -c '{"function":"addMember","Args":["skyhuihui","lion","123456","A","1","0.1","0.2","ASDASDSA"]}' 0.1是上级分成比例，0.2是账户自身交易比例
	if function == "addMember" {
		return t.addMember(stub, args)
	} else if function == "delMember" {
		return t.delMember(stub, args) //删除成员（没有考虑删除成员之后货币的总量如何变化）peer chaincode invoke -C mychannel -n mycc -c '{"function":"delMember","Args":["skyhuihui","lion"]}' -o orderer.example.com:7050
	} else if function == "initLedger" {
		return t.initLedger(stub, args) //生成管理员账号peer chaincode invoke -C mychannel -n mycc -c '{"function":"initLedger","Args":[]}' -o orderer.example.com:7050
	} else if function == "queryByName" {
		return t.queryByName(stub, args) //根据名字返回成员信息 peer chaincode invoke -C mychannel -n mycc -c '{"function":"queryByName","Args":["skyhuihui"]}'
	} else if function == "initCurrency" {
		return t.initCurrency(stub, args) //创建代币peer chaincode invoke -C mychannel -n mycc -c '{"function":"initCurrency","Args":["Netkiller Token","NKC","1000000","skyhuihui"]}'
	} else if function == "showToken" {
		return t.showToken(stub, args) //展示所有的代币信息 peer chaincode invoke -C mychannel -n mycc -c '{"function":"showToken","Args":[]}'
	} else if function == "showTokenUser" {
		return t.showTokenUser(stub, args) //展示某个代币的所有信息peer chaincode invoke -C mychannel -n mycc -c '{"function":"showTokenUser","Args":["NKC"]}'
	} else if function == "transferToken" {
		return t.transferToken(stub, args) //转账 peer chaincode invoke -C mychannel -n mycc -c '{"function":"transferToken","Args":["skyhuihui","lion","NKC","12.584"]}'
	} else if function == "frozenAccount" {
		return t.frozenAccount(stub, args) // 冻结某个账户,该账户冻结之后无法转账peer chaincode invoke -C mychannel -n mycc -c '{"function":"frozenAccount","Args":["lion","true","skyhuihui"]}'
	} else if function == "tokenHistory" {
		return t.tokenHistory(stub, args) // 查询某个币的交易记录 peer chaincode invoke -C mychannel -n mycc -c '{"function":"tokenHistory","Args":["NKC"]}'
	} else if function == "userTokenHistory" {
		return t.userTokenHistory(stub, args) //查询某个币某个用户的记录 peer chaincode invoke -C mychannel -n mycc -c '{"function":"userTokenHistory","Args":["NKC","lion"]}'
	} else if function == "getHistoryForKey" {
		return t.getHistoryForKey(stub, args) //查询某个键的所有交易记录 peer chaincode invoke -C mychannel -n mycc -c '{"function":"getHistoryForKey","Args":["lion"]}'
	} else if function == "burnToken" {
		return t.burnToken(stub, args) //回收某个账户的特定代币数量 peer chaincode invoke -C mychannel -n mycc -c '{"function":"burnToken","Args":["NKC","5000","123","skyhuihui"]}'
	} else if function == "mintToken" {
		return t.mintToken(stub, args) //增加某个币的数量，peer chaincode invoke -C mychannel -n mycc -c '{"function":"mintToken","Args":["NKC","5000","skyhuihui"]}'
	} else if function == "changefee" {
		return t.changefee(stub, args) //修改某个会员的交易比例 //peer chaincode invoke -C mychannel -n mycc -c '{"function":"changefee","Args":["skyhuihui","lion","0.3"]}'
	} else if function =="changeDeviceId"{
		return t.changeDeviceId(stub,args)//更新设备信息 //peer chaincode invoke -C mychannel -n mycc -c '{"function":"changeDeviceId","Args":["lion","asdasasdsd"]}'
	}

	return shim.Error("Invalid function name，input correct funciton name.")
}
func delmenber(stub shim.ChaincodeStubInterface, member Member) bool {
	//msg:=fmt.Sprintf("会员删除失败，%s",memberName)
	err := stub.DelState(member.MemberName)
	if err != nil {
		return false
	}
	return true
}

//保存member
func Putmember(stub shim.ChaincodeStubInterface, member Member) ([]byte, bool) {
	b, err := json.Marshal(member)
	if err != nil {
		return nil, false
	}
	// 保存member状态
	//err = stub.PutState(string(member.MemberID), b)
	//if err != nil {
	//	return nil, false
	//}
	err = stub.PutState(member.MemberName, b)
	if err != nil {
		return nil, false
	}
	return b, true
}

// 根据会员ID查询信息状态
// args: MemberID
//func GetIDInfo(stub shim.ChaincodeStubInterface, MemberID string) (Member, bool) {
//	var member Member
//	// 根据会员ID查询信息状态
//	b, err := stub.GetState(MemberID)
//	if err != nil {
//		return member, false
//	}
//	if b == nil {
//		return member, false
//	}
//	// 对查询到的状态进行反序列化
//	err = json.Unmarshal(b, &member)
//	if err != nil {
//		return member, false
//	}
//	// 返回结果
//	return member, true
//}
//根据会员名字查询信息的调用功能
func GetNameInfo(stub shim.ChaincodeStubInterface, MemberName string) (Member, bool) {
	var member Member
	// 根据会员名字查询信息状态
	b, err := stub.GetState(MemberName)
	if err != nil {
		return member, false
	}
	if b == nil {
		return member, false
	}
	// 对查询到的状态进行反序列化
	err = json.Unmarshal(b, &member)
	if err != nil {
		return member, false
	}
	// 返回结果
	return member, true
}

//根据用户名查询会员信息
func (t *SmartContract) queryByName(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	name := args[0]
	// 根据会员名字查询信息状态
	b, err := stub.GetState(name)
	if err != nil {
		return shim.Error("failed to find")
	}
	if b == nil {
		return shim.Error("The account is empty")
	}

	return shim.Success(b)

}
//根据用户名更新设备码
//peer chaincode invoke -C mychannel -n mycc -c '{"function":"changeDeviceId","Args":["lion","asdasasdsd"]}'
func (t *SmartContract) changeDeviceId(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	name := args[0]
	Deviceid:=args[1]

	member,ok:=GetNameInfo(stub,name)
	if !ok{
		return shim.Error("the MemberName is empty")
	}
    member.DeviceID=Deviceid
    b,tag:=Putmember(stub,member)
    if !tag{
    	return shim.Error("failed to update the DeviceId of the member")
	}
	return shim.Success(b)
}

//修改用户的交易手续费，暂时不考虑修改上级的手续费分成
//peer chaincode invoke -C mychannel -n mycc -c '{"function":"changefee","Args":["skyhuihui","lion","0.3"]}'
func (t *SmartContract) changefee(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	Operator := args[0]
	//操作者的信息
	a, ok := GetNameInfo(stub, Operator)
	if !ok {
		return shim.Error("Cannot find operator")
	} else if a.MemberClass != Admin {
		return shim.Error("You should enough privilege to do")
	}
	membername := args[1]
	var member Member
	member, exist := GetNameInfo(stub, membername)
	if !exist {
		return shim.Error("failed to delete")
	}
	fee, _ := strconv.ParseFloat(args[2], 64)
	member.Fee = fee //修改现有的会员的手续费比例，然后将会员信息上链
	memberAsbyte, tag := Putmember(stub, member)
	if !tag {
		return shim.Error("failed to update Member fee")
	}
	return shim.Success(memberAsbyte)

}
func (t *SmartContract) initLedger(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	name := "skyhuihui"
	member := Member{
		MemberID:      0,
		MemberName:    name,
		MemberPwd:     Pwd,
		MemberClass:   Admin,
		Frozen:        false,
		DeviceID:     DeviceID,
		BalanceOf:     map[string]float64{},
		Charge:        map[string]float64{},
		SuperiorCount: map[string]float64{},
		Fee:           0, //管理员不存在手续费，该转多少是多少

	}
	re, ok := Putmember(stub, member)
	if !ok {
		return shim.Error("failed to create administrator")
	}
	err := stub.SetEvent("InitLedger", []byte{})
	if err != nil {
		return shim.Error(err.Error())
	}
	ans := "initLedger success," + string(re)
	return shim.Success([]byte(ans))
}

//删除会员
func (t *SmartContract) delMember(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	Operator := args[0]
	a, ok := GetNameInfo(stub, Operator)
	if !ok {
		return shim.Error("Cannot find operator")
	} else if a.MemberClass != Admin {
		return shim.Error("You should enough privilege to do")
	}
	Delmembername := args[1]
	var member Member
	member, exist := GetNameInfo(stub, Delmembername)
	if !exist {
		return shim.Error("failed to delete")
	}
	tag := delmenber(stub, member)
	if !tag {
		return shim.Error("failed to delete")

	}
	return shim.Success(nil)
}

//添加会员,增加了参数 开户享有分成的比例 以及这个被创建账户交易产生的比例
//peer chaincode invoke -C mychannel -n mycc -c '{"function":"addMember","Args":["skyhuihui","lion","123456","A","1"，"0.1","0.2","Aasdasdas”,"safecode"]}'
func (t *SmartContract) addMember(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//当前操作者
	if len(args) != 9 {
		return shim.Error("Incorrect number of arguments. Expecting 9")
	}
	//查询当前操作者的身份信息、并确认待添加会员姓名是否冲突

	currOperatorName := args[0]
	//操作者的信息
	currOperator, ok := GetNameInfo(stub, currOperatorName)
	if !ok {
		return shim.Error("Nnvalid operator name")
	}

	level, err := strconv.Atoi(args[4])
	if err != nil {
		return shim.Error(err.Error())
	}
	scale, err := strconv.ParseFloat(args[5], 64) //上级能获得分成的比例
	if err != nil {
		return shim.Error(err.Error())
	}
	radio, err := strconv.ParseFloat(args[6], 64) //被创建账户交易时候的手续费或者能获得奖励的比例
	if err != nil {
		return shim.Error(err.Error())
	}
	supcount := map[string]float64{}
	supcount[currOperatorName] = scale

	rand.Seed(time.Now().Unix())
	id := rand.Intn(10000000) + 1
	member := Member{
		MemberID:      id,
		MemberName:    args[1],
		MemberPwd:     args[2],
		MemberClass:   args[3],
		MemberLevel:   level,
		DeviceID:      args[7],
		Frozen:        false,
		BalanceOf:     map[string]float64{},
		Charge:        map[string]float64{},
		SuperiorCount: supcount,
		Fee:           radio,
		SafeCode:      args[8],
	}
	fmt.Println("member:", member)
	//返回会员名字的信息
	_, exist := GetNameInfo(stub, member.MemberName)
	if exist {
		return shim.Error("The member name has existed")
	}
	//权限判断
	isAuthorized := false
	if member.MemberClass == "A" {
		if member.MemberLevel == 1 {
			if currOperator.MemberClass == "Admin" {
				member.SuperiorCount[currOperatorName] = 0 //管理员创建A1直接100%享有它的售卖的手续费，这里设置为0，后面更新后台账户资金，后台直接获得手续费，不用扣除上级。
				isAuthorized = true
			}
		}
		if member.MemberLevel == 2 {
			if currOperator.MemberClass == "A" && currOperator.MemberLevel == 1 {
				isAuthorized = true
			}
		}
		if member.MemberLevel == 3 {
			if currOperator.MemberClass == "A" && currOperator.MemberLevel == 2 {
				isAuthorized = true
			}
		}
	} else if member.MemberClass == "B" {
		if member.MemberLevel == 1 {
			if currOperator.MemberClass == "Admin" {
				member.SuperiorCount[currOperatorName] = 0 //管理员创建B1,默认为0，对称，实际B1这个值没用
				isAuthorized = true
			}
		}
		if member.MemberLevel == 2 {
			if currOperator.MemberClass == "B" && currOperator.MemberLevel == 1 {
				isAuthorized = true
			}
		}
		if member.MemberLevel == 3 {
			if currOperator.MemberClass == "B" && currOperator.MemberLevel == 2 {
				isAuthorized = true
			}
		}
		if member.MemberLevel == 4 {
			if currOperator.MemberClass == "B" && currOperator.MemberLevel == 3 {
				isAuthorized = true
			}
		}
	} else if member.MemberClass == "C" {
		member.SuperiorCount[currOperatorName] = 0 //扫码注册的C类，管理员直接100%享有它的售卖的手续费
		member.MemberLevel = 0                     //C会员没有等级，默认统一设置为0
		isAuthorized = true
	}
	//如果权限满足
	if isAuthorized {
		memberbytes, b1 := Putmember(stub, member)
		if !b1 {
			return shim.Error("b保存信息失败")
		}
		return shim.Success(memberbytes)
	} else {
		fmt.Println("添加会员失败！权限不满足，当前会员等级" + currOperator.MemberClass + string(currOperator.MemberLevel) + "，待添加会员等级" + member.MemberClass + string(member.MemberLevel))
		return shim.Error("Error, Authorized Fail !")
	}
	return shim.Success(nil)
}

//判断会员是否具有该货币
func isCurrency(member Member, _currency string) bool {

	if _, ok := member.BalanceOf[_currency]; !ok {
		return false
	}
	return true
}

//转账  管理员转账，一律视为激活各等级用户代币交易功能。
func Transfer(stub shim.ChaincodeStubInterface, _from Member, _to Member, _currency string, _value float64, sc float64, bonus float64) ([]byte, bool) {

	var rev []byte

	if _from.Frozen {
		msg := "From 账号冻结"
		rev, _ = json.Marshal(msg)
		return rev, false
	}
	if _to.Frozen {
		msg := "To 账号冻结"
		rev, _ = json.Marshal(msg)
		return rev, false

	}
	//这里特别定义，只要转账者携带该货币，即使被转账者没有携带也可以转账，直接让被转账者生成新货币,建议生成新货币统一用后台来开
	if !isCurrency(_from, _currency) {
		msg := "货币符号不存在"
		rev, _ = json.Marshal(msg)
		return rev, false
	}
	tokenAsbytes, err := stub.GetState(TokenKey)

	if err != nil {
		msg := "获取货币信息失败"
		rev, _ = json.Marshal(msg)
		return rev, false
	}
	token := Token{}
	err = json.Unmarshal(tokenAsbytes, &token)
	if err != nil {
		msg := "反序列失败"
		rev, _ = json.Marshal(msg)
		return rev, false
	}
	if _from.BalanceOf[_currency] >= _value+sc { //转正者钱必须大于转账金额和手续费才能转
		_from.BalanceOf[_currency] = _from.BalanceOf[_currency] - _value - sc //扣除转账费用和手续费
		_to.BalanceOf[_currency] = _to.BalanceOf[_currency] + bonus + _value
		//激活B类奖励冻结功能,
		if _to.MemberClass == "B" {
			if _, exit := _to.Charge[_currency]; !exit {
				_to.Charge[_currency] = 0 //第一次激活的时候，该值需要设定为0
			}
		}
		//买家是B类才会有奖励分成，才会需要交易解冻，管理员转账开户不算
		if _to.MemberClass == "B" && _from.MemberClass != Admin {
			FrozenCharge(stub, _to, _currency, _value)
		}

		//更新转账和被转账者的信息
		_, ok := Putmember(stub, _from)
		if !ok {
			msg := "更新转账者信息失败"
			rev, _ = json.Marshal(msg)
			return rev, false
		}
		_, tag := Putmember(stub, _to)
		if !tag {
			msg := "更新接受者信息失败"
			rev, _ = json.Marshal(msg)
			return rev, false
		}
		//如果本次交易手续费和奖励抵消，不用更新，否则更新管理员账户，后续只需要考虑上级分成
		if sc-bonus != 0 {
			admin, good := GetNameInfo(stub, AdminName)
			if !good {
				msg := "获取管理员信息失败"
				rev, _ = json.Marshal(msg)
				return rev, false
			}
			admin.BalanceOf[_currency] = admin.BalanceOf[_currency] + sc - bonus
			_, g := Putmember(stub, admin)
			if !g {
				msg := "更新管理员信息失败"
				rev, _ = json.Marshal(msg)
				return rev, false
			}
		}

		//更新token的用户记录
		//token.Currency[_currency].User[_from.MemberName] = _from.BalanceOf[_currency]
		//token.Currency[_currency].User[_to.MemberName] = _to.BalanceOf[_currency]

		//将交易记录纳入代币当中，存放在区块链上
		TransferRecord := TransactionRecord{_from.MemberName, _to.MemberName, _currency, _value, stub.GetTxID()}
		recordList := make([]TransactionRecord, 0)
		recordList = append(token.Currency[_currency].Record, TransferRecord)
		var cur Currency
		cur = Currency{token.Currency[_currency].TokenName, token.Currency[_currency].TokenSymbol, token.Currency[_currency].TotalSupply, recordList}
		token.Currency[_currency] = cur
		tokenAsBytes2, err := json.Marshal(token)
		if err != nil {
			msg := "序列化代币信息失败"
			rev, _ = json.Marshal(msg)
			return rev, false
		}
		err = stub.PutState(TokenKey, tokenAsBytes2)
		if err != nil {
			msg := "更新代币信息失败"
			rev, _ = json.Marshal(msg)
			return rev, false
		}

		msg := "success to transfer"
		rev, _ = json.Marshal(msg)
		return rev, true
	} else {
		msg := "账户余额不足"
		rev, _ = json.Marshal(msg)
		return rev, false
	}
}

//解冻买币者获得的手续费分成
func FrozenCharge(stub shim.ChaincodeStubInterface, to Member, _currency string, _value float64) {

	//买币家解冻
	if to.Charge[_currency] >= _value { //奖励分成比此次交易多
		to.BalanceOf[_currency] += _value //奖励的钱入账
		to.Charge[_currency] -= _value    //未解冻的奖励钱减少
	} else {
		to.BalanceOf[_currency] += to.Charge[_currency] //手续费分成比此次交易少，全部解封
		to.Charge[_currency] = 0                        //未解冻的奖励钱清零
	}
}

//Token键，值是所有以token为名，值为currency的键值对
func (t *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {

	token := &Token{Currency: map[string]Currency{}}

	tokenAsBytes, err := json.Marshal(token)
	err = stub.PutState(TokenKey, tokenAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	} else {
		fmt.Printf("Init Token %s \n", string(tokenAsBytes))
	}
	err = stub.SetEvent("tokenInvoke", []byte{})
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}
func (t *SmartContract) showToken(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	tokenAsBytes, err := stub.GetState(TokenKey)
	if err != nil {
		return shim.Error(err.Error())
	} else {
		fmt.Printf("GetState(%s)) %s \n", TokenKey, string(tokenAsBytes))
	}
	return shim.Success(tokenAsBytes)
}
func (t *SmartContract) showTokenUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	_token := args[0]
	token := Token{}
	existAsBytes, err := stub.GetState(TokenKey)
	if err != nil {
		return shim.Error(err.Error())
	} else {
		fmt.Printf("GetState(%s)) %s \n", TokenKey, string(existAsBytes))
	}
	json.Unmarshal(existAsBytes, &token)
	if _, ok := token.Currency[_token]; !ok {
		return shim.Error("The Token doesn't exist")
	} //1.3版本增加的部分，判断查询代币是否存在
	reToekn, err := json.Marshal(token.Currency[_token])
	if err != nil {
		return shim.Error(err.Error())
	} else {
		fmt.Printf("Account balance %s \n", string(reToekn))
	}
	return shim.Success(reToekn)
}

//创建代币peer chaincode invoke -C mychannel -n mycc -c '{"function":"initCurrency","Args":["Netkiller Token","NKC","1000000","skyhuihui"]}
func (t *SmartContract) initCurrency(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}
	operator := args[3]
	member, exist := GetNameInfo(stub, operator)
	if !exist {
		return shim.Error("The administrator account is empty")
	} else if member.MemberClass != Admin {
		return shim.Error("You should enough privilege to do")
	}

	_name := args[0]
	_symbol := args[1]
	_supply, _ := strconv.ParseFloat(args[2], 64)
	token := Token{}
	existAsBytes, err := stub.GetState(TokenKey)
	if err != nil {
		return shim.Error(err.Error())
	} else {
		fmt.Printf("GetState(%s)) %s \n", TokenKey, string(existAsBytes))
	}
	json.Unmarshal(existAsBytes, &token)
	if _, ok := token.Currency[_symbol]; ok {
		return shim.Error("Token has been created")
	}
	//user := make(map[string]float64)
	//user[_account] = _supply
	//token.Currency[_symbol] = Currency{TokenName: _name, TokenSymbol: _symbol, TotalSupply: _supply,User: user,Record:[]TransactionRecord{}}
	token.Currency[_symbol] = Currency{TokenName: _name, TokenSymbol: _symbol, TotalSupply: _supply, Record: []TransactionRecord{}}
	tokenAsBytes, _ := json.Marshal(token)
	err = stub.PutState(TokenKey, tokenAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	} else {
		fmt.Printf("Init Token %s \n", string(tokenAsBytes))
	}
	member.BalanceOf[_symbol] = _supply
	//member.Charge[_symbol]=0 //初始化货币的时候，某货币的手续费余额为0
	_, ok := Putmember(stub, member)
	if !ok {
		return shim.Error("failed to put member")
	}

	err = stub.SetEvent("tokenInvoke", []byte{})
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(tokenAsBytes)
}

//货币交易 管理员转账，一律视为激活各等级用户代币交易功能。
//转账 peer chaincode invoke -C mychannel -n mycc -c '{"function":"transferToken","Args":["skyhuihui","lion","NKC","12.584"]}'
func (t *SmartContract) transferToken(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}
	_from := args[0]
	_to := args[1]
	_currency := args[2]
	_amount, _ := strconv.ParseFloat(args[3], 64)
	if _amount <= 0 {
		return shim.Error("Incorrect number of amount")
	}
	memberfrom, exist := GetNameInfo(stub, _from)
	if !exist {
		return shim.Error("Invalid user to transfer")
	}
	fromradio := memberfrom.Fee //转账者手续费的比例

	//fromtop:=memberfrom.SuperiorCount //获得上级账户名字，以及上级能够享有的比例

	memberto, ok := GetNameInfo(stub, _to)
	if !ok {
		return shim.Error("Invaild user to receive")
	}
	toradio := memberto.Fee //接受者获得奖励的比例

	//totop:=memberto.SuperiorCount //获得上级账户名字，以及上级能够享有的比例

	var sc float64    //提前设置一个手续费变量
	var bonus float64 //提前设置一个奖励变量

	//转账是否正确
	isAuthorized := false
	//A出售BC有手续费存在
	if memberfrom.MemberClass == "A" {
		sc = _amount * fromradio //A交易产生的手续费
		if memberto.MemberClass == "B" {
			bonus = _amount * toradio //B购买获得的奖励
			isAuthorized = true
		} else if memberto.MemberClass == "C" {
			isAuthorized = true
		}
	} else if memberfrom.MemberClass == "B" {
		if memberto.MemberClass == "C" || memberto.MemberClass == "A" {
			isAuthorized = true
		}
	} else if memberfrom.MemberClass == "C" {
		if memberto.MemberClass == "B" {
			sc = _amount * fromradio  //C出售产生的手续费
			bonus = _amount * toradio //B购买获得的奖励
			isAuthorized = true
		}
	} else if memberfrom.MemberClass == Admin {
		if memberto.MemberClass == "A" || memberto.MemberClass == "B" || memberto.MemberClass == "C" {
			isAuthorized = true
		}
	}
	if isAuthorized {

		result, ok := Transfer(stub, memberfrom, memberto, _currency, _amount, sc, bonus)
		if !ok {
			return shim.Error(string(result))
		}
		//卖家上层账户获得分成
		if memberfrom.MemberClass == "A" && memberfrom.MemberLevel != 1 { //管理员不需要计算分成，A1和B1也没有上级分成。

			var adminfee float64 //管理员账户需要增加的费用
			for memberfrom.MemberLevel != 1 {
				m, charge, x := FromProfit(stub, memberfrom, _currency, sc)
				if !x {
					return shim.Error("The top of the fromaccount failed to update")
				}
				memberfrom = m
				sc = charge
				adminfee += charge
			}
			// 更新管理员账户也就是后台账户信息,非A1才会更新
			admin, z := GetNameInfo(stub, AdminName)
			if !z {
				return shim.Error("failed to get Admin")
			}
			admin.BalanceOf[_currency] = admin.BalanceOf[_currency] - adminfee //把上级手续费分成全部分发完
			_, t := Putmember(stub, admin)
			if !t {
				return shim.Error("failed to update Admin")
			}
		}
		//买家上层账户获得分成 只有B类才做这个操作
		if memberto.MemberClass == "B" && memberto.MemberLevel != 1 && memberfrom.MemberClass != Admin {
			var adminfee float64
			for memberto.MemberLevel != 1 {
				m, charge, x := ToProfit(stub, memberto, _currency, bonus)
				if !x {
					return shim.Error("The top of the toaccount failed to update")
				}
				memberto = m
				bonus = charge
				adminfee += charge
			}
			// 更新管理员账户也就是后台账户信息,非B1才会更新
			admin, z := GetNameInfo(stub, AdminName)
			if !z {
				return shim.Error("failed to get Admin")
			}
			admin.BalanceOf[_currency] = admin.BalanceOf[_currency] - adminfee //把上级奖励全部分发完
			_, t := Putmember(stub, admin)
			if !t {
				return shim.Error("failed to update Admin")
			}
		}
		err := stub.SetEvent("tokenInvoke", []byte{})
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(result)

	} else {
		return shim.Error("Error, Authorized Fail !")
	}

	return shim.Success(nil)
}
func FromProfit(stub shim.ChaincodeStubInterface, from Member, _currency string, sc float64) (Member, float64, bool) {

	var memberfromtop Member
	var fromscale float64
	fromtop := from.SuperiorCount
	for key1, v1 := range fromtop {
		member1, ok := GetNameInfo(stub, key1)
		if !ok {
			return member1, 0, false
		}
		memberfromtop = member1
		fromscale = v1
	}
	//卖家上级获得手续费分成
	a := sc * fromscale
	memberfromtop.BalanceOf[_currency] += a //非B类获得的分成直接用
	_, x := Putmember(stub, memberfromtop)
	if !x {
		return memberfromtop, 0, false
	}
	return memberfromtop, a, true
}
func ToProfit(stub shim.ChaincodeStubInterface, to Member, _currency string, bonus float64) (Member, float64, bool) {

	var membertotop Member
	var toscale float64
	totop := to.SuperiorCount
	for key1, v1 := range totop {
		member1, ok := GetNameInfo(stub, key1)
		if !ok {
			return member1, 0, false
		}
		membertotop = member1
		toscale = v1
	}
	//买家上级获得奖励
	a := bonus * toscale
	membertotop.Charge[_currency] += a //B类获得分成，需要设置，此为下级获得的冻结金额,其他类默认null
	_, x := Putmember(stub, membertotop)
	if !x {
		return membertotop, 0, false
	}
	return membertotop, a, true
}

//冻结账户
func (t *SmartContract) frozenAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	_account := args[0]
	_status := args[1]
	Operator := args[2]
	a, ok := GetNameInfo(stub, Operator)
	if !ok {
		return shim.Error("Cannot find operator")
	} else if a.MemberClass != Admin {
		return shim.Error("You should get enough privilege to do")
	}

	member, exist := GetNameInfo(stub, _account)
	if !exist {
		return shim.Error("Cannot find member")
	}

	var status bool
	if _status == "true" {
		status = true
	} else {
		status = false
	}

	member.Frozen = status
	b, tag := Putmember(stub, member)
	if !tag {
		return shim.Error("Failed to change frozen ")
	} else {
		fmt.Printf("frozenAccount - end %s \n", b)
	}

	err := stub.SetEvent("tokenInvoke", []byte{})
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

//获取代币交易记录
func (t *SmartContract) tokenHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	_currency := args[0]

	tokenAsBytes, err := stub.GetState(TokenKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	token := Token{}
	json.Unmarshal(tokenAsBytes, &token)
	resultAsBytes, _ := json.Marshal(token.Currency[_currency].Record)

	fmt.Printf("Token Record %s \n", string(resultAsBytes))
	return shim.Success(resultAsBytes)
}

//获取某个用户某个代币交易记录
func (t *SmartContract) userTokenHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	_currency := args[0]
	_account := args[1]

	tokenAsBytes, err := stub.GetState(TokenKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	token := Token{}
	json.Unmarshal(tokenAsBytes, &token)
	var userRecord []TransactionRecord
	index := 0
	for k, v := range token.Currency[_currency].Record {
		if token.Currency[_currency].Record[k].From == _account || token.Currency[_currency].Record[k].To == _account {
			userRecord = append(userRecord, v)
			index++
		}
	}

	resultAsBytes, _ := json.Marshal(userRecord)
	fmt.Printf("Token Record nums %d \n", index)
	fmt.Printf("Token Record  %s \n", string(resultAsBytes))
	return shim.Success(resultAsBytes)
}

func (t *SmartContract) getHistoryForKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	marbleId := args[0]

	// 返回某个键的所有历史值
	resultsIterator, err := stub.GetHistoryForKey(marbleId)
	if err != nil {
		return shim.Error(err.Error())
	}

	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResult.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		//时间戳格式化
		txtimestamp := queryResult.Timestamp
		tm := time.Unix(txtimestamp.Seconds, 0)
		timeString := tm.Format("2006-01-02 03:04:05 PM")
		buffer.WriteString(timeString)
		buffer.WriteString("\"")

		buffer.WriteString("{\"Value\":")
		buffer.WriteString("\"")
		buffer.WriteString(string(queryResult.Value))
		buffer.WriteString("\"")

		buffer.WriteString("{\"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(queryResult.IsDelete))
		buffer.WriteString("\"")

		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	fmt.Printf("- getMarblesByRange queryResult:\n%s\n", buffer.String())
	return shim.Success(buffer.Bytes())
}

func (t *SmartContract) burnToken(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}
	_currency := args[0]
	_amount, _ := strconv.ParseFloat(args[1], 64)
	_account := args[2]
	operator := args[3]
	member, exist := GetNameInfo(stub, operator)
	if !exist {
		return shim.Error("The administrator account is empty")
	} else if member.MemberClass != Admin {
		return shim.Error("You should enough privilege to do")
	}

	burnmember, ok := GetNameInfo(stub, _account)
	if !ok {
		return shim.Error("the destroyed token account is empty")
	}
	tag := isCurrency(burnmember, _currency)
	if !tag {
		return shim.Error("the currency doesn't exist")
	}

	tokenAsBytes, err := stub.GetState(TokenKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Printf("Token before %s \n", string(tokenAsBytes))

	token := Token{}

	json.Unmarshal(tokenAsBytes, &token)
	//member的代币要减少，总token的代币也要减少
	if burnmember.BalanceOf[_currency] >= _amount {
		cur := token.Currency[_currency]
		cur.TotalSupply -= _amount //货币销毁，所以这里货币总量要减少
		token.Currency[_currency] = cur
		burnmember.BalanceOf[_currency] -= _amount
		fmt.Println("success to recycle")
		//更新burnmenber
		a, ok := Putmember(stub, burnmember)
		if !ok {
			return shim.Error("failed to update member ")
		} else {
			fmt.Println("Admin after:", a)
		}
		//更新代币
		tokenAsBytes, err = json.Marshal(token)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = stub.PutState(TokenKey, tokenAsBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
		fmt.Printf("Token after %s \n", string(tokenAsBytes))
		fmt.Printf("burnToken %s \n", string(tokenAsBytes))
	} else {
		return shim.Error("burnmember's token is not enough to decrease")
	}
	err = stub.SetEvent("tokenInvoke", []byte{})
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SmartContract) mintToken(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	_currency := args[0]
	_amount, _ := strconv.ParseFloat(args[1], 64)
	_account := args[2]
	member, exist := GetNameInfo(stub, _account)
	if !exist {
		return shim.Error("The administrator account is empty")
	} else if member.MemberClass != Admin {
		return shim.Error("You should enough privilege to do")
	}
	ok := isCurrency(member, _currency)
	if !ok {
		return shim.Error("the destroyed token account is empty")
	}

	tokenAsBytes, err := stub.GetState(TokenKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Printf("Token before %s \n", string(tokenAsBytes))
	token := Token{}
	json.Unmarshal(tokenAsBytes, &token)
	//代币更新，管理员代币更新
	cur := token.Currency[_currency]
	cur.TotalSupply += _amount //增发导致总量增加
	token.Currency[_currency] = cur
	member.BalanceOf[_currency] += _amount //更新管理员代币
	fmt.Println("success to increase")
	//更新menber
	a, ok := Putmember(stub, member)
	if !ok {
		return shim.Error("failed to update member ")
	} else {
		fmt.Println("Admin after : ", a)
	}
	//更新代币
	tokenAsBytes, err = json.Marshal(token)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(TokenKey, tokenAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Printf("Token after %s \n", string(tokenAsBytes))
	fmt.Printf("minToken %s \n", string(tokenAsBytes))

	err = stub.SetEvent("tokenInvoke", []byte{})
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

//Go语言的入口Main函数
func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %", err)
	}
}
