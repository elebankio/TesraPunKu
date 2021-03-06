package dosnode

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
	"net/http"
	"os"
	"strconv"

	"github.com/DOSNetwork/core/onchain"
)

func (d *DosNode) startRESTServer() (err error) {
	defer fmt.Println("End RESTServer")
	mux := http.NewServeMux()
	mux.HandleFunc("/", d.status)
	mux.HandleFunc("/balance", d.balance)
	mux.HandleFunc("/setGasPrice", d.setGasPrice)
	mux.HandleFunc("/setGasLimit", d.setGasLimit)
	mux.HandleFunc("/enableAdmin", d.enableAdmin)
	mux.HandleFunc("/disableAdmin", d.disableAdmin)
	mux.HandleFunc("/enableGuardian", d.enableGuardian)
	mux.HandleFunc("/disableGuardian", d.disableGuardian)
	mux.HandleFunc("/signalGroupFormation", d.signalGroupFormation)
	mux.HandleFunc("/signalGroupDissolve", d.signalGroupDissolve)
	mux.HandleFunc("/signalBootstrap", d.signalBootstrap)
	mux.HandleFunc("/signalRandom", d.signalRandom)
	mux.HandleFunc("/dkgTest", d.dkgTest)
	d.handleGroupFormation()
	s := http.Server{Addr: ":8080", Handler: mux}
	go func() {
		<-d.ctx.Done()
		s.Shutdown(context.Background())
	}()
	err = s.ListenAndServe()
	return
}

func (d *DosNode) status(w http.ResponseWriter, r *http.Request) {
	isPendingNode, err := d.chain.IsPendingNode(d.id)
	if err != nil {
		html := "err : " + err.Error() + "\n|"
		w.Write([]byte(html))
		return
	}
	html := "=================================================" + "\n|"
	html = html + "Version           : " + d.config.VERSION + "\n|"
	html = html + "StartTime         : " + d.startTime.Format("2006-01-02T15:04:05.999999-07:00") + "\n|"
	html = html + "Address           : " + fmt.Sprintf("%x", d.p.GetID()) + "\n|"
	html = html + "IP                : " + fmt.Sprintf("%s", d.p.GetIP()) + "\n|"
	html = html + "NumOfMembers      : " + strconv.Itoa(d.p.NumOfMembers()) + "\n|"
	html = html + "State             : " + d.state + "\n|"
	html = html + "IsPendingNode     : " + strconv.FormatBool(isPendingNode) + "\n|"
	html = html + "TotalQuery        : " + strconv.Itoa(d.totalQuery) + "\n|"
	html = html + "FulfilledQuery    : " + strconv.Itoa(d.fulfilledQuery) + "\n|"
	html = html + "Group Number      : " + strconv.Itoa(d.numOfworkingGroup) + "\n|"
	html = html + "gasPrice          : " + strconv.FormatUint(d.chain.GetGasPrice(), 10) + "\n|"
	html = html + "gasLimit          : " + strconv.FormatUint(d.chain.GetGasLimit(), 10) + "\n|"
	balance, err := d.chain.Balance()
	if err != nil {
		html = html + "Balance           : " + err.Error() + "\n|"
	} else {
		html = html + "Balance           : " + balance.String() + "\n|"
	}
	workingGroupNum, err := d.chain.GetWorkingGroupSize()
	if err != nil {
		html = html + "WorkingGroupSize  : " + err.Error() + "\n|"
	} else {
		html = html + "WorkingGroupSize  : " + strconv.FormatUint(workingGroupNum, 10) + "\n|"
	}
	expiredGroupNum, err := d.chain.GetExpiredWorkingGroupSize()
	if err != nil {
		html = html + "ExpiredGroupSize  : " + err.Error() + "\n|"
	} else {
		html = html + "ExpiredGroupSize  : " + strconv.FormatUint(expiredGroupNum, 10) + "\n|"
	}
	pendingGroupNum, err := d.chain.NumPendingGroups()
	if err != nil {
		html = html + "PendingGroupSize  : " + err.Error() + "\n|"
	} else {
		html = html + "PendingGroupSize  : " + strconv.FormatUint(pendingGroupNum, 10) + "\n|"
	}
	pendingNodeNum, err := d.chain.NumPendingNodes()
	if err != nil {
		html = html + "PendingNodeSize   : " + err.Error() + "\n|"
	} else {
		html = html + "PendingNodeSize   : " + strconv.FormatUint(pendingNodeNum, 10) + "\n|"
	}
	curBlk, err := d.chain.CurrentBlock()
	if err != nil {
		html = html + "CurrentBlock      : " + err.Error() + "\n"
	} else {
		html = html + "CurrentBlock      : " + strconv.FormatUint(curBlk, 10) + "\n"
	}
	html = html + "=================================================" + "\n"
	w.Write([]byte(html))
}

func (d *DosNode) balance(w http.ResponseWriter, r *http.Request) {
	html := "Balance :"
	result, err := d.chain.Balance()
	if err != nil {
		html = html + err.Error()
	} else {
		html = html + result.String()
	}
	w.Write([]byte(html))
}

func (d *DosNode) setGasLimit(w http.ResponseWriter, r *http.Request) {
	gasLimits, ok := r.URL.Query()["gasLimit"]
	if !ok || len(gasLimits) == 0 {
		return
	}
	g, ok := new(big.Int).SetString(gasLimits[0], 10)
	if !ok {
		d.logger.Error(fmt.Errorf("GasLimit SetString error"))
		return
	}
	if g.Cmp(big.NewInt(0)) == 0 {
		d.logger.Error(fmt.Errorf("GasLimit cannot be set to 0"))
		return
	}
	d.logger.Info(fmt.Sprintf("Set GasLimit to %v", g))
	d.chain.SetGasLimit(g)
}

func (d *DosNode) setGasPrice(w http.ResponseWriter, r *http.Request) {
	gasPrices, ok := r.URL.Query()["gasPrice"]
	if !ok || len(gasPrices) == 0 {
		return
	}
	g, ok := new(big.Int).SetString(gasPrices[0], 10)
	if !ok {
		d.logger.Error(fmt.Errorf("GasPrice SetString error"))
		return
	}
	// Set to 0 means using estimate gas price by default
	d.logger.Info(fmt.Sprintf("Set GasPrice to %v wei", g))
	d.chain.SetGasPrice(g)
}

func (d *DosNode) enableAdmin(w http.ResponseWriter, r *http.Request) {
	d.logger.Info("isAdmin")
	d.isAdmin = true
}

func (d *DosNode) disableAdmin(w http.ResponseWriter, r *http.Request) {
	d.logger.Info("disable admin")
	d.isAdmin = false
}

func (d *DosNode) enableGuardian(w http.ResponseWriter, r *http.Request) {
	d.logger.Info("enableGuardian")
	d.isGuardian = true
}

func (d *DosNode) disableGuardian(w http.ResponseWriter, r *http.Request) {
	d.logger.Info("disable guardian")
	d.isGuardian = false
}

func (d *DosNode) signalBootstrap(w http.ResponseWriter, r *http.Request) {
	cid := -1
	switch r.Method {
	case "GET":
		for k, v := range r.URL.Query() {
			fmt.Printf(" %s: %s\n", k, v)
			if k == "cid" {
				i, err := strconv.Atoi(v[0])
				if err == nil && i >= 0 {
					cid = i
				}
			}
		}
	default:
	}
	if cid >= 0 {
		d.chain.SignalBootstrap(big.NewInt(int64(cid)))
	}
}

func (d *DosNode) signalRandom(w http.ResponseWriter, r *http.Request) {
	d.chain.SignalRandom()
}

func (d *DosNode) signalGroupFormation(w http.ResponseWriter, r *http.Request) {
	d.chain.SignalGroupFormation()
}

func (d *DosNode) signalGroupDissolve(w http.ResponseWriter, r *http.Request) {
	d.chain.SignalGroupDissolve()
}

func (d *DosNode) p2pTest(w http.ResponseWriter, r *http.Request) {
	d.End()
}
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	buffer := make([]byte, 40)

	for {
		bytesread, err := file.Read(buffer)

		if err != nil {
			break
		}
		lines = append(lines, string(buffer[:bytesread]))
	}
	return lines, nil
}
func (d *DosNode) dkgTest(w http.ResponseWriter, r *http.Request) {
	d.logger.Info("dkgTest start  ")
	groupID := big.NewInt(0)
	start := -1
	end := -1
	switch r.Method {
	case "GET":
		for k, v := range r.URL.Query() {
			fmt.Printf(" %s: %s\n", k, v)
			if k == "start" {
				i, err := strconv.Atoi(v[0])
				if err == nil && i >= 0 {
					start = i
				}
			} else if k == "end" {
				i, err := strconv.Atoi(v[0])
				if err == nil && i >= 0 {
					end = i
				}
			} else if k == "gid" {
				i, err := strconv.Atoi(v[0])
				if err == nil && i >= 0 {
					groupID = groupID.SetInt64(int64(i))
				}
			}
		}
	}
	d.logger.Info("start=" + strconv.Itoa(start))
	d.logger.Info("end=" + strconv.Itoa(end))
	d.logger.Info("gid=" + groupID.String())
	//members, err := readLines("/home/goProject/src/github.com/DOSNetwork/core/vault/members.txt")
	//if err != nil {
	//	d.logger.Error(err)
	//	return
	//}
	var members [3]string
	members[0] = "0x78E05d5BfF1Cb316Dc78BCa387aD34dB63299799"
	members[1] = "0xF7FaA65023fB73D99AEc4d15f46bC28a3cBf4Cad"
	members[2] = "0x957fDeCb162601cc8b9864f6b020799e78B5d40E"
	if start >= 0 && end >= 0 {
		if len(members) < (end - start) {
			d.logger.Info(fmt.Sprintf("members size not enough: %d", len(members)))
			return
		}
		var participants [][]byte

		for i := start; i < end; i++ {
			d.logger.Info(string(i) + ":" + members[i])
			temp, _ := hexutil.Decode(members[i])
			participants = append(participants, temp)
		}

		d.onchainEvent <- &onchain.LogGrouping{
			GroupId: groupID,
			NodeId:  participants,
		}
	}
}
func (d *DosNode) queryTest(w http.ResponseWriter, r *http.Request) {
	groupId := big.NewInt(0)
	lastSys := big.NewInt(0)
	userSeed := big.NewInt(0)
	requestId := big.NewInt(0)
	switch r.Method {
	case "GET":
		for k, v := range r.URL.Query() {
			fmt.Printf("%s: %s\n", k, v)
			if k == "gid" {
				i, err := strconv.Atoi(v[0])
				if err == nil && i >= 0 {
					groupId = groupId.SetInt64(int64(i))
				}
			} else if k == "rid" {
				i, err := strconv.Atoi(v[0])
				if err == nil && i >= 0 {
					requestId = requestId.SetInt64(int64(i))
				}
			}
		}
	}
	d.onchainEvent <- &onchain.LogRequestUserRandom{
		RequestId:            requestId,
		LastSystemRandomness: lastSys,
		UserSeed:             userSeed,
		DispatchedGroupId:    groupId,
	}
}
