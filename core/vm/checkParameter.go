package vm

import (
	"fmt"
	"time"
	"bytes"
	"errors"
	"math/big"
	"crypto/sha256"

	"golang.org/x/crypto/ripemd160"
	"github.com/GenaroNetwork/Genaro-Core/core/types"
	"github.com/GenaroNetwork/Genaro-Core/common"
	"github.com/GenaroNetwork/Genaro-Core/crypto"
	"github.com/GenaroNetwork/Genaro-Core/common/hexutil"
	"strings"
	"github.com/GenaroNetwork/Genaro-Core/params"
)


func isSpecialAddress(address common.Address,optionTxMemorySize uint64) bool {
	for _, v := range common.SpecialAddressList {
		if bytes.Compare(address.Bytes(), v.Bytes()) == 0{
			return  true
		}
	}
	dist := address.Sub(common.OptionTxBeginSaveAddress)
	if dist>=0 && dist<int64(optionTxMemorySize) {
		return true
	}
	return false
}

func CheckSpecialTxTypeSyncSidechainStatusParameter( s types.SpecialTxInput,caller common.Address, genaroConfig *params.GenaroConfig) error {
	if true == isSpecialAddress(s.SpecialTxTypeMortgageInit.FromAccount,genaroConfig.OptionTxMemorySize) {
		return errors.New("fromAccount error")
	}

	if caller !=  common.OfficialAddress {
		return errors.New("caller address of this transaction is not invalid")
	}

	if 64 != len(s.SpecialTxTypeMortgageInit.Dataversion) {
		return errors.New("Parameter Dataversion  error")
	}

	if 64 != len(s.SpecialTxTypeMortgageInit.FileID) {
		return errors.New("Parameter fileID  error")
	}
	if 20 != len(s.SpecialTxTypeMortgageInit.FromAccount) {
		return errors.New("Parameter fromAccount  error")
	}
	if 1 < len(s.SpecialTxTypeMortgageInit.Sidechain) {
		for k,v := range s.SpecialTxTypeMortgageInit.Sidechain{
			if 20 != len(k) {
				return errors.New("Parameter mortgage account  error")
			}
			if v.ToInt().Cmp(big.NewInt(0)) < 0 {
				return errors.New("Parameter Sidechain")
			}
		}
	} else {
		return errors.New("Parameter side chain length less than zero")
	}
	return nil
}


func CheckspecialTxTypeMortgageInitParameter( s types.SpecialTxInput,caller common.Address) error {
	var tmp  big.Int
	timeLimit := s.SpecialTxTypeMortgageInit.TimeLimit.ToInt()
	tmp.Mul(timeLimit,big.NewInt(86400))
	endTime :=  tmp.Add(&tmp,big.NewInt(s.SpecialTxTypeMortgageInit.CreateTime)).Int64()
	if s.SpecialTxTypeMortgageInit.CreateTime > s.SpecialTxTypeMortgageInit.EndTime ||
		s.SpecialTxTypeMortgageInit.CreateTime > time.Now().Unix() ||
		s.SpecialTxTypeMortgageInit.EndTime != endTime {
		return errors.New("Parameter CreateTime or EndTime  error")
	}
	if caller != s.SpecialTxTypeMortgageInit.FromAccount {
		return errors.New("Parameter FromAccount  error")
	}
	if len(s.SpecialTxTypeMortgageInit.FileID) != 64 {
		return errors.New("Parameter FileID  error")
	}
	mortgageTable := s.SpecialTxTypeMortgageInit.MortgageTable
	authorityTable := s.SpecialTxTypeMortgageInit.AuthorityTable
	if len(authorityTable) != len(mortgageTable) {
		return errors.New("Parameter authorityTable != mortgageTable  error")
	}
	for k,v := range authorityTable {
		if v < 0 || v > 3 {
			return errors.New("Parameter authority type  error")
		}
		if mortgageTable[k].ToInt().Cmp(big.NewInt(0)) < 0 {
			return errors.New("Parameter mortgage amount is less than zero")
		}
	}
	return nil
}

func CheckSynchronizeShareKeyParameter( s types.SpecialTxInput, genaroConfig *params.GenaroConfig) error {

	if true == isSpecialAddress(s.SynchronizeShareKey.RecipientAddress,genaroConfig.OptionTxMemorySize) {
		return errors.New("update  chain SynchronizeShareKey fail")
	}

	if len(s.SynchronizeShareKey.ShareKeyId) != 64 {
		return errors.New("Parameter ShareKeyId  error")
	}
	if len(s.SynchronizeShareKey.ShareKey) > 0 {
		return errors.New("Parameter ShareKey  error")
	}
	if s.SynchronizeShareKey.Shareprice.ToInt().Cmp(big.NewInt(0)) < 0 {
		return errors.New("Parameter Shareprice  is less than zero")
	}
	return nil
}

func CheckUnlockSharedKeyParameter( s types.SpecialTxInput) error {
	if len(s.SynchronizeShareKey.ShareKeyId) != 64 {
		return errors.New("Parameter ShareKeyId  error")
	}
	return nil
}

func CheckStakeTx(s types.SpecialTxInput, state StateDB, genaroConfig *params.GenaroConfig) error {
	adress := common.HexToAddress(s.Address)
	if isSpecialAddress(adress,genaroConfig.OptionTxMemorySize){
		return errors.New("param [address] can't be special address")
	}

	genaroPrice := state.GetGenaroPrice()
	if s.Stake < genaroPrice.MinStake {
		return errors.New("value of stake must larger than MinStake")
	}

	// 判断是否已经申请了退注
	if state.IsAlreadyBackStake(adress) {
		return errors.New("account in back stake list")
	}
	return nil
}

func CheckSyncHeftTx(caller common.Address, s types.SpecialTxInput, genaroConfig *params.GenaroConfig) error {
	if caller !=  common.OfficialAddress {
		return errors.New("caller address of this transaction is not invalid")
	}

	adress := common.HexToAddress(s.Address)
	if isSpecialAddress(adress,genaroConfig.OptionTxMemorySize){
		return errors.New("param [address] can't be special address")
	}

	if s.Heft <= 0 {
		return errors.New("value of heft must larger than zero")
	}

	return nil
}

func CheckApplyBucketTx(s types.SpecialTxInput,genaroConfig *params.GenaroConfig) error {
	adress := common.HexToAddress(s.Address)
	if isSpecialAddress(adress,genaroConfig.OptionTxMemorySize){
		return errors.New("param [address] can't be special address")
	}

	for _, v := range s.Buckets {
		if len(v.BucketId) != 64 {
			return errors.New("length of bucketId must be 64")
		}
	}
	return nil
}

func CheckTrafficTx(s types.SpecialTxInput, genaroConfig *params.GenaroConfig) error {
	adress := common.HexToAddress(s.Address)
	if isSpecialAddress(adress,genaroConfig.OptionTxMemorySize){
		return errors.New("param [address] can't be special address")
	}

	if s.Traffic <= 0 {
		errors.New("value of traffic must larger than zero")
	}
	return nil
}


func CheckSyncNodeTx(caller common.Address, s types.SpecialTxInput, db StateDB) error {
	stake, _ := db.GetStake(caller)
	existNodes := db.GetStorageNodes(caller)
	stakeVlauePerNode := db.GetStakePerNodePrice()

	if len(s.NodeID) == 0 {
		return errors.New("length of nodeId must larger then 0")
	}

	paramAddress := common.HexToAddress(s.Address)
	//caller和节点待绑定账户是否一致
	if caller != paramAddress {
		return errors.New("address in param is not equal with callerAddress of this Tx")
	}


	//校验节点是否已经绑定过
	if db.GetAddressByNode(s.NodeID) != "" {
		return errors.New("the input node have been bound by themselves or others")
	}

	// 验证节点绑定签名
	// 拼接message
	msg := s.NodeID + s.Address

	sig, err := hexutil.Decode(s.Sign)
	if err != nil {
		return errors.New("sign without 0x prefix")
	}

	recoveredPub, err := crypto.SigToPub(crypto.Keccak256([]byte(msg)), sig)
	if err != nil {
		return errors.New("ECRecover error when valid sign")
	}

	//get publickey
	pubKey := crypto.CompressPubkey(recoveredPub)

	//log.Info(fmt.Sprintf("publicKey:%x", pubKey))
	genNodeID := generateNodeId(pubKey)
	//log.Info(fmt.Sprintf("genNodeId:%s", genNodeID))
	//log.Info(fmt.Sprintf("s.nodeId:%s", s.NodeID))
	if genNodeID != s.NodeID {
		return errors.New("sign valid error, nodeId mismatch")
	}

	var nodeNum int = 1
	if existNodes != nil {
		nodeNum += len(existNodes)
	}

	needStakeVale := big.NewInt(0)
	needStakeVale.Mul(big.NewInt(int64(nodeNum)), stakeVlauePerNode)

	currentStake := new(big.Int).Mul(new(big.Int).SetUint64(stake), common.BaseCompany)

	//log.Info(fmt.Sprintf("currentStake:%s", currentStake.String()))
	//log.Info(fmt.Sprintf("needStakeVale:%s", needStakeVale.String()))

	if needStakeVale.Cmp(currentStake) == 1 {
		return errors.New("none enough stake to synchronize node")
	}
	return nil
}


func generateNodeId(b []byte) string {
	sha256byte := sha256.Sum256(b)
	ripemder := ripemd160.New()
	ripemder.Write(sha256byte[:])
	hashBytes := ripemder.Sum(nil)
	nodeId := fmt.Sprintf("%x", hashBytes)
	return nodeId
}

func CheckPunishmentTx(caller common.Address,s types.SpecialTxInput, genaroConfig *params.GenaroConfig) error {
	adress := common.HexToAddress(s.Address)
	if isSpecialAddress(adress,genaroConfig.OptionTxMemorySize){
		return errors.New("param [address] can't be special address")
	}

	if caller !=  common.OfficialAddress {
		return errors.New("caller address of this transaction is not invalid")
	}
	return nil
}

func CheckBackStakeTx(caller common.Address, state StateDB) error {
	ok,backStakeList := state.GetAlreadyBackStakeList()
	if !ok {
		return errors.New("userBackStake fail")
	}
	genaroPrice := state.GetGenaroPrice()
	if len(backStakeList) > int(genaroPrice.BackStackListMax) {
		return errors.New("BackStackList too long")
	}
	// 判断是否是绑定用户
	if state.IsBindingAccount(caller) {
		return errors.New("account is binding")
	}
	// 判断是否已经申请了退注
	if state.IsAlreadyBackStake(caller) {
		return errors.New("account in back stake list")
	}
	// 判断账号是否在禁止退注的名单中
	if state.IsAccountExistInForbidBackStakeList(caller) {
		return errors.New("account in forbid backstake list")
	}
	return nil
}

func CheckSynStateTx(caller common.Address, state StateDB) error {
	genaroPrice := state.GetGenaroPrice()
	synStateAccount := common.HexToAddress(genaroPrice.SynStateAccount)
	if caller != synStateAccount {
		return errors.New("caller address of this transaction is not invalid")
	}
	return nil
}

func CheckSyncFileSharePublicKeyTx(s types.SpecialTxInput, genaroConfig *params.GenaroConfig) error {
	adress := common.HexToAddress(s.Address)
	if isSpecialAddress(adress,genaroConfig.OptionTxMemorySize){
		return errors.New("param [address] can't be special address")
	}

	if s.FileSharePublicKey == "" {
		return errors.New("public key for file share can't be null")
	}
	return nil
}

func CheckPriceRegulation(caller common.Address ,s types.SpecialTxInput) error {
	if caller !=  common.GenaroPriceAddress {
		return errors.New("caller address of this transaction is not invalid")
	}

	if s.StakeValuePerNode == nil && s.BucketApplyGasPerGPerDay == nil && s.TrafficApplyGasPerG == nil && s.OneDayMortgageGes == nil && s.OneDaySyncLogGsaCost == nil {
		return errors.New("none price to update")
	}

	return nil
}

func CheckUnbindNodeTx(caller common.Address,s types.SpecialTxInput, existNodes []string) error{
	if existNodes == nil {
		return errors.New("none node of this account need to unbind")
	}

	if s.NodeID == "" {
		return errors.New("nodeId is null")
	}

	for _, v := range existNodes{
		if v == s.NodeID {
			return nil
		}
	}
	return errors.New("this node does not belong to this account")
}

// 账号绑定检查
func CheckAccountBindingTx(caller common.Address,s types.SpecialTxInput, state StateDB) error{
	// 检查是否是官方账号
	if caller !=  common.OfficialAddress {
		return errors.New("caller address of this transaction is not invalid")
	}
	// 主账号
	mainAccount := common.HexToAddress(s.Address)
	// 子账号
	subAccount := common.HexToAddress(s.Message)
	if bytes.EqualFold(mainAccount.Bytes(),subAccount.Bytes()) {
		return errors.New("same account")
	}
	// 主账号是否是候选者
	if !state.IsCandidateExist(mainAccount) {
		return errors.New("mainAddr is not a candidate")
	}
	// 主账号绑定数量是否超出限制
	genaroPrice := state.GetGenaroPrice()
	if state.GetSubAccountsCount(mainAccount) > int(genaroPrice.MaxBinding) {
		return errors.New("binding enough")
	}
	// 绑定的子账号是否已经是一个主账号
	if state.IsBindingMainAccount(subAccount) {
		return errors.New("sub account is a main account")
	}
	// 子账号是否是候选者或存在于子账号队列中
	thisMainAccount := state.GetMainAccount(subAccount)
	if !state.IsCandidateExist(subAccount) && thisMainAccount == nil{
		return errors.New("subAddr is not a candidate")
	}
	// 账号是否已经处于绑定状态
	if thisMainAccount != nil && bytes.Compare(thisMainAccount.Bytes(),mainAccount.Bytes()) == 0 {
		return errors.New("has binding")
	}

	return nil
}

// 检查输入参数，并返回执行类型
// 1 主账号解绑
// 2 子账号解绑
// 3 主账号解绑子账号
func CheckAccountCancelBindingTx(caller common.Address,s types.SpecialTxInput, state StateDB) (t int,err error){
	// 判断账号类型
	if state.IsBindingMainAccount(caller) {
		if strings.EqualFold(s.Address,"") {
			t = 1
		} else {
			subAccount := common.HexToAddress(s.Address)
			if state.IsBindingSubAccount(subAccount) {
				thisMainAccount := state.GetMainAccount(subAccount)
				if thisMainAccount !=nil && bytes.EqualFold(thisMainAccount.Bytes(),caller.Bytes()){
					t = 3
				} else {
					err = errors.New("not binding account")
				}
			}else {
				err = errors.New("not binding account")
			}
		}

	} else if state.IsBindingSubAccount(caller) {
		t = 2
	} else {
		err = errors.New("not binding account")
	}
	return
}

// 添加禁止退注名单的交易检查
func CheckAddAccountInForbidBackStakeListTx(caller common.Address,s types.SpecialTxInput, state StateDB) error{
	// 检查是否是官方账号
	if caller !=  common.OfficialAddress {
		return errors.New("caller address of this transaction is not invalid")
	}
	account := common.HexToAddress(s.Address)
	// 检查账号是否有押注
	stake,err := state.GetStake(account)
	if err != nil {
		return err
	}
	if stake == 0 {
		return errors.New("account stake is zero")
	}
	// 判断是否已经在禁止名单中
	if state.IsAccountExistInForbidBackStakeList(account) {
		return errors.New("account is in forbid list")
	}
	return nil
}

// 移除退注账号禁止名单的检查
func CheckDelAccountInForbidBackStakeListTx(caller common.Address,s types.SpecialTxInput, state StateDB) error {
	// 检查是否是官方账号
	if caller !=  common.OfficialAddress {
		return errors.New("caller address of this transaction is not invalid")
	}
	account := common.HexToAddress(s.Address)
	// 检查账号是否在禁止名单中
	ok := state.IsAccountExistInForbidBackStakeList(account)
	if !ok {
		return errors.New("account is not in forbid list")
	}
	return nil
}

// 设置全局变量交易参数检查
func CheckSetGlobalVar(caller common.Address,s types.SpecialTxInput) error {
	// 检查是否是官方账号
	if caller !=  common.OfficialAddress {
		return errors.New("caller address of this transaction is not invalid")
	}

	if s.RatioPerYear >= 100 || s.CoinRewardsRatio >= 100 || s.StorageRewardsRatio >= 100{
		return errors.New("Ratio is not invalid")
	}

	return nil
}

// 增加币池的检查
func CheckAddCoinpool(caller common.Address,s types.SpecialTxInput, state StateDB) error {
	balance := state.GetBalance(caller)
	if s.AddCoin.ToInt().Cmp(big.NewInt(0)) <= 0 {
		return errors.New("Value is not invalid")
	}
	if balance.Cmp(s.AddCoin.ToInt()) < 0 {
		return errors.New("Balance is not enough")
	}
	return nil
}

