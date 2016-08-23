package main

import (
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"

	"strconv"
	"time"
	"math"
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

	RequestId  string `json:"request_id"`
	LenderId string `json:"lender_id"`
	LendeeId string `json:"lendee_id"`
	LatestHistoryId string `json:"latest_history_id"`
}

type Item struct {

	RequestId  string `json:"request_id"`
	ItemId  string `json:"item_id"`
	ItemName string `json:"item_name"`
}


type RequestHistory struct {

	RequestId  string `json:"request_id"`
	HistoryId  string `json:"history_id"`
	ChangerId string `json:"changer_id"`
	StatusFrom string `json:"status_from"`
	StatusTo string `json:"status_to"`
	TimeStamp string `json:"time_stamps"`
	Note string `json:"note"`
}



//==============================================================================================================================
//	Init Function - Called when the user deploys the chaincode
//==============================================================================================================================
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
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

		if !t.validate_timestamp(timestamp) {
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

	if function == "get_customer" {

		if function == "get_customer" {

			if len(args) != 2 {
				fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed")
			}

			//customer_id := args[0]
			//receiver_id := args[1]
			//
			//return t.get_customer(stub, customer_id, receiver_id)

		}
	}
	return nil, errors.New("QUERY: No such function.")

}

// state changing function should;
//  - receive timestamp as an argument (to keep consistency between validating peers)
//  - should validate the received timestamp with
//  -- it is within 5min of the machine time
//  -- it is larger than the previous state

func (t *SimpleChaincode) create_request(stub *shim.ChaincodeStub, lender_id string, lendee_id string, timestamp string, items []string) ([]byte, error) {

	return nil, nil
}



func (t *SimpleChaincode) validate_timestamp(timestamp string) (bool) {
	time_on_timestamp , err := time.Parse(time.RFC3339Nano, timestamp)
	//time_on_timestamp , err := time.Parse(time.RFC3339Nano, "2013-06-05T14:1043.678Z")
	if err == nil && math.Abs(time_on_timestamp.Sub(time.Now()).Minutes()) <= MAX_TIME_DIFF_IN_MIN {
		return true
	}
	return false
}