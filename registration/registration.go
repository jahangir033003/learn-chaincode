
package main

import (
	"errors"
	"fmt"
	//"strconv"
	//"encoding/json"


	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"reflect"
	"strings"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type User struct {
	UserID        	string       	`json:"UserID,omitempty"`
	FirstName 	string 		`json:"UserID,omitempty"`
	LastName 	string 		`json:"UserID,omitempty"`
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Init called, initializing chaincode")

	return nil, nil
}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// Handle different functions
	if function == "createUser" {
		// create assetID
		return t.createUser(stub, args)
	}/* else if function == "updateUser" {
		// create assetID
		return t.updateUser(stub, args)
	} else if function == "deleteUser" {
		// Deletes an asset by ID from the ledger
		return t.deleteUser(stub, args)
	}*/
	return nil, errors.New("Received unknown invocation: " + function)
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.readUser(stub, args)
	}
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}



func (t *SimpleChaincode) createUser(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	_,erval:=t. createOrUpdateUser(stub, args)
	return nil, erval
}



func (t *SimpleChaincode) updateUser(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	_,erval:=t. createOrUpdateUser(stub, args)
	return nil, erval
}

/*
func (t *SimpleChaincode) deleteUser(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var assetID string // asset ID
	var err error
	var stateIn AssetState

	// validate input data for number of args, Unmarshaling to asset state and obtain asset id
	stateIn, err = t.validateInput(args)
	if err != nil {
		return nil, err
	}
	assetID = *stateIn.AssetID
	// Delete the key / asset from the ledger
	err = stub.DelState(assetID)
	if err != nil {
		err = errors.New("DELSTATE failed! : "+ fmt.Sprint(err))
		return nil, err
	}
	return nil, nil
}
*/
func (t *SimpleChaincode) readUser(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var userID string
	var state User = User{}
	var err error

	userID = strings.TrimSpace(args[0])
	// Get the state from the ledger
	assetBytes, err:= stub.GetState(userID)
	if err != nil  || len(assetBytes) ==0{
		err = errors.New("Unable to get user state from ledger")
		return nil, err
	}
	err = json.Unmarshal(assetBytes, &state)
	if err != nil {
		err = errors.New("Unable to unmarshal state data obtained from ledger")
		return nil, err
	}
	return assetBytes, nil
}


func (t *SimpleChaincode) validateInput(args []string) (stateIn User, err error) {
	var userID string
	var state User = User{}

	if len(args) !=1 {
		err = errors.New("Incorrect number of arguments. Expecting a JSON strings with mandatory assetID")
		return state, err
	}
	jsonData:=args[0]
	userID = ""
	stateJSON := []byte(jsonData)
	err = json.Unmarshal(stateJSON, &stateIn)
	if err != nil {
		err = errors.New("Unable to unmarshal input JSON data")
		return state, err
		// state is an empty instance of asset state
	}
	// was userId present?
	// The nil check is required because the asset id is a pointer.
	// If no value comes in from the json input string, the values are set to nil

	if stateIn.UserID != "" {
		userID = strings.TrimSpace(stateIn.UserID)
		if userID == ""{
			err = errors.New("UserId not passed")
			return state, err
		}
	} else {
		err = errors.New("User id is mandatory in the input JSON data")
		return state, err
	}


	stateIn.UserID = userID
	return stateIn, nil
}
//******************** createOrUpdateAsset ********************/

func (t *SimpleChaincode) createOrUpdateUser(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var userID string
	var err error
	var userIn User
	var userStub User


	// validate input data for number of args, Unmarshaling to asset state and obtain asset id

	userIn, err = t.validateInput(args)
	if err != nil {
		return nil, err
	}
	userID = userIn.UserID
	// Partial updates introduced here
	// Check if asset record existed in stub
	assetBytes, err:= stub.GetState(userID)
	if err != nil || len(assetBytes)==0{
		// This implies that this is a 'create' scenario
		userStub = userIn // The record that goes into the stub is the one that cme in
	} else {
		// This is an update scenario
		err = json.Unmarshal(assetBytes, &userStub)
		if err != nil {
			err = errors.New("Unable to unmarshal JSON data from stub")
			return nil, err
			// state is an empty instance of asset state
		}
		// Merge partial state updates
		userStub, err =t.mergePartialState(userStub,userIn)
		if err != nil {
			err = errors.New("Unable to merge state")
			return nil,err
		}
	}
	stateJSON, err := json.Marshal(userStub)
	if err != nil {
		return nil, errors.New("Marshal failed for contract state" + fmt.Sprint(err))
	}
	// Get existing state from the stub


	// Write the new state to the ledger
	err = stub.PutState(userID, stateJSON)
	if err != nil {
		err = errors.New("PUT ledger state failed: "+ fmt.Sprint(err))
		return nil, err
	}
	return nil, nil
}

func (t *SimpleChaincode) mergePartialState(oldState User, newState User) (User,  error) {

	old := reflect.ValueOf(&oldState).Elem()
	new := reflect.ValueOf(&newState).Elem()
	for i := 0; i < old.NumField(); i++ {
		oldOne:=old.Field(i)
		newOne:=new.Field(i)
		if ! reflect.ValueOf(newOne.Interface()).IsNil() {
			oldOne.Set(reflect.Value(newOne))
		}
	}
	return oldState, nil
}
