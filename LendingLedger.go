package main

import (
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"

	"strconv"
	"time"
	"math"
	"strings"
	"encoding/json"
)

type  SimpleChaincode struct {
}

const (
	REQUEST_CREATED = iota //  initial state, items can only be added at this state.
	ITEMS_SHIPPED_TO_LENDER = iota //  lender has sent the item
	ITEMS_RECEIVED_BY_LENDER = iota //  lendee has received the item
	ITEMS_SHIPPED_BY_LENDER = iota //  lendee has sent back the item
	ITEMS_RECEIVED_BY_LENDEE = iota //  lender has received the item back
)

const MAX_TIME_DIFF_IN_MIN = 5 // to validate timestamp

type Request struct {
	RequestId  string `json:"request_id"`  // 00001
	LenderId string `json:"lender_id"`
	LendeeId string `json:"lendee_id"`
	LatestHistoryId string `json:"latest_history_id"`
}

type Item struct { // ITEM/00001/00001

	RequestId  string `json:"request_id"`  // 00001
	ItemId  string `json:"item_id"`   // 00001
	ItemName string `json:"item_name"`
}


type RequestHistory struct { // REQ_HIST/00001/00001

	RequestId  string `json:"request_id"` // 00001
	HistoryId  string `json:"history_id"`  // 00001
	//ChangerId string `json:"changer_id"`
	StatusFrom int `json:"status_from"`
	StatusTo int `json:"status_to"`
	TimeStamp string `json:"time_stamp"`
	Note string `json:"note"`
}


// 以下結果返し用
type RequestSet struct{
	Requests []RequestRecord `json:"requests"`
}

type RequestRecord struct{
	RequestId  string `json:"request_id"`
	LenderId string `json:"lender_id"`
	LendeeId string `json:"lendee_id"`
	LatestHistoryId string `json:"latest_history_id"`
	UpdatedTimeStamp string `json:"updated_timestamp"`
	Items []Item `json:"items"`
	Status string `json:"status"`
}


//==============================================================================================================================
//	Init Function - Called when the user deploys the chaincode
//==============================================================================================================================
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	// init request_id counter
	err := stub.PutState("REC_CTR", []byte("00000"))
	if err != nil {
		return nil, errors.New("Unable to put the state")
	}

	return nil, nil
}

//==============================================================================================================================
//	 Router Functions
//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//		  initial arguments passed to other things for use in the called function e.g. name -> ecert
//==============================================================================================================================
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	//caller, caller_affiliation, err := t.get_caller_data(stub)

	//if err != nil { return nil, errors.New("Error retrieving caller information")}


	if function == "create_request" {

		if len(args) < 5 {
			fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("INVOKE: Incorrect number of arguments passed")
		}

		lender_id := args[0]
		lendee_id := args[1]
		timestamp := args[2]

		if !validate_timestamp(timestamp) {
			fmt.Printf("Timestamp: Incorrect format or too far from system clock "+timestamp); return nil, errors.New("INVOKE: Timestamp: Incorrect format or too far from system clock"+timestamp)
		}

		item_counts, err := strconv.ParseInt(args[3],10,64)

		if err !=nil || int64(len(args)) != int64(4) + item_counts {
			fmt.Printf("item_counts and len(args) doesn't match"); return nil, errors.New("INVOKE: item_counts and len(args) doesn't match")
		}

		items := args[4:4 + item_counts ]

		return t.create_request(stub, lender_id, lendee_id, timestamp, items)

	}
	return nil, errors.New("Function of that name doesn't exist.")
}
//=================================================================================================================================
//	Query - Called on chaincode query. Takes a function name passed and calls that function. Passes the
//  		initial arguments passed are passed on to the called function.
//=================================================================================================================================
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	if function == "get_all_requests" {


		if len(args) != 0 {
			fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed")
		}

		//customer_id := args[0]
		//receiver_id := args[1]
		//
		return t.get_all_requests(stub)


	}
	return nil, errors.New("QUERY: No such function.")

}

// state changing function should;
//  - receive timestamp as an argument (to keep consistency between validating peers)
//  - should validate the received timestamp with
//  -- it is within 5min of the machine time => validate_timestamp
//  -- it is larger than the previous state

func (t *SimpleChaincode) create_request(stub *shim.ChaincodeStub, lender_id string, lendee_id string, timestamp string, items []string) ([]byte, error) {

	// Register Request
	request_id := next_req_ctr(stub)
	reqKey := "REQ/"+request_id
	request:= Request{
		RequestId:request_id,
		LenderId:lender_id,
		LendeeId:lendee_id,
		LatestHistoryId:"00000",
	}
	bytes, err := json.Marshal(request)
	if err != nil {
		return nil, errors.New("Error creating Reqest record")
	}
	err = stub.PutState( reqKey, []byte(bytes))
	if err != nil {
		return nil, errors.New("Unable to put the state")
	}

	// Register Items
	for i, itemName := range items{
		s:=strings.Repeat("0",5)+ string(i)
		l:=len(s)
		item_id:=s[l-5:l-1] // 00000
		itemKey:="ITEM/"+request_id+"/"+ item_id
		item:=Item{
			RequestId:request_id,
			ItemId:item_id,
			ItemName:itemName,
		}
		bytes, err := json.Marshal(item)
		if err != nil {
			return nil, errors.New("Error creating Reqest record")
		}
		err = stub.PutState( itemKey, []byte(bytes))
		if err != nil {
			return nil, errors.New("Unable to put the state")
		}
	}

	// Register RequestHistory
	hist:=RequestHistory{
		RequestId:request_id,
		HistoryId:request.LatestHistoryId, //00000
		StatusFrom:REQUEST_CREATED,
		StatusTo:ITEMS_SHIPPED_TO_LENDER,
		TimeStamp:timestamp,
		Note:"",
	}
	histKey := "REQ_HIST/"+hist.HistoryId+"/"+hist.HistoryId
	histBytes, err := json.Marshal(hist)
	if err != nil {
		return nil, errors.New("Error creating Reqest record")
	}
	err = stub.PutState( histKey, []byte(histBytes))
	if err != nil {
		return nil, errors.New("Unable to put the state")
	}

	increment_req_ctr(stub)
	return nil, nil
}


func (t *SimpleChaincode) get_all_requests(stub *shim.ChaincodeStub) ([]byte, error) {

	var rset RequestSet
	var req Request
	var record RequestRecord
	var items []Item
	var item Item
	var hist RequestHistory

	reqsIter, err := stub.RangeQueryState("REQ/", "REQ/~")
	if err != nil {
		return nil, errors.New("Unable to start the iterator")
	}

	defer reqsIter.Close()

	for reqsIter.HasNext() {
		_, valAsbytes, iterErr := reqsIter.Next()
		if iterErr != nil {
			return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
		}

		if err = json.Unmarshal(valAsbytes, req) ; err != nil {
			return nil, errors.New("Error unmarshalling data "+string(valAsbytes))
		}

		itemsIter, err := stub.RangeQueryState("ITEM/"+req.RequestId+"/", "ITEM/"+req.RequestId+"/~")
		if err != nil {
			return nil, errors.New("Unable to start the iterator")
		}
		for itemsIter.HasNext() {
			_, itemAsbytes, err := itemsIter.Next()
			if err != nil {	return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)	}
			if err = json.Unmarshal(itemAsbytes, item) ; err != nil { return nil, errors.New("Error unmarshalling data "+string(itemAsbytes))}
			items = append(items,item)
		}
		itemsIter.Close()

		statusKey:="REQ_HIST/"+req.RequestId+"/"+req.LatestHistoryId
		histAsbytes, err := stub.GetState(statusKey)
		if err != nil {return nil, errors.New("Error getting customer data of "+statusKey)}
		if err = json.Unmarshal(histAsbytes, hist) ; err != nil {return nil, errors.New("Error unmarshalling data "+string(histAsbytes))}


		record = RequestRecord{
			RequestId:req.RequestId ,
			LenderId:req.LenderId ,
			LendeeId:req.LendeeId ,
			LatestHistoryId:req.LatestHistoryId ,
			UpdatedTimeStamp:hist.TimeStamp,
			Items:items  ,
			Status:status_in_string(hist.StatusTo)  ,
		}

		rset.Requests = append(rset.Requests,record)
	}

	bytes, err := json.Marshal(rset.Requests)
	if err != nil {
		return nil, errors.New("Error creating Reqests record")
	}
	return []byte(bytes), nil
}




func validate_timestamp(timestamp string) (bool) {
	time_on_timestamp , err := time.Parse(time.RFC3339Nano, timestamp)
	//time_on_timestamp , err := time.Parse(time.RFC3339Nano, "2013-06-05T14:1043.678Z")
	if err == nil && math.Abs(time_on_timestamp.Sub(time.Now()).Minutes()) <= MAX_TIME_DIFF_IN_MIN {
		return true
	}
	return false
}

func next_req_ctr(stub *shim.ChaincodeStub) (string) {
	next_ctr, err := stub.GetState("REQ_CTR")
	if err != nil || len(next_ctr) == 0 {
		panic("Failed to GetState")
	}
	ctr, err := strconv.Atoi(string(next_ctr) )
	if err != nil {panic(err)}
	if ctr > 60000  {
		panic("Counter overflow.")
	}

	str:=strings.Repeat("0",5)+ string(ctr + 1)
	l:=len(str)
	return str[l-5:l-1]
}

func increment_req_ctr(stub *shim.ChaincodeStub) () {
	next_ctr, err := stub.GetState("REQ_CTR")
	if err != nil || len(next_ctr) == 0 {
		panic("Failed to GetState")
	}
	ctr, err := strconv.Atoi(string(next_ctr) )
	if err != nil {panic(err)}
	stub.PutState("REQ_CTR",[]byte(string(ctr + 1)))
	return
}

func status_in_string(status int) (string) {
	var val = ""
	switch status{
		case REQUEST_CREATED : val = "REQUEST_CREATED" //  initial state, items can only be added at this state.
		case ITEMS_SHIPPED_TO_LENDER: val = "ITEMS_SHIPPED_TO_LENDER" //  lender has sent the item
		case ITEMS_RECEIVED_BY_LENDER : val = "ITEMS_RECEIVED_BY_LENDER"  //  lendee has received the item
		case ITEMS_SHIPPED_BY_LENDER : val = "ITEMS_SHIPPED_BY_LENDER" //  lendee has sent back the item
		case ITEMS_RECEIVED_BY_LENDEE : val = "ITEMS_RECEIVED_BY_LENDEE"  //  lender has received the item back
	}
	return val
}

//=================================================================================================================================
//	 Main - main - Starts up the chaincode
//=================================================================================================================================
func main() {

	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Chaincode: %s", err)
	}
}
