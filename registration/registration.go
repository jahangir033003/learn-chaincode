
package main

import (
	"errors"
	"fmt"
	//"strconv"
	//"encoding/json"


	"github.com/hyperledger/fabric/core/chaincode/shim"
	//"github.com/imdario/mergo"
	"encoding/json"
	"reflect"
	"strings"

	"time"
	"strconv"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type User struct {
	UserID        	*string       	`json:"UserID,omitempty"`
	FirstName 	*string 	`json:"FirstName,omitempty"`
	LastName 	*string 	`json:"LastName,omitempty"`
	Email 		*string 	`json:"Email,omitempty"`
	Password 	*string 	`json:"Password,omitempty"`
	Gender		*uint8		`json:"Gender,omitempty"` // 1= Male, 2 = Female, 3 = Others
	Dcoument	*string		`json:"Dcoument,omitempty"`
	CreatedDate	*time.Time	`json:"CreatedDate,omitempty"`
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Init called, initializing chaincode")

	return nil, nil
}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// Handle different functions
	if function == "createUser" {
		// create User
		return t.createUser(stub, args)
	} else if function == "updateUser" {
		// create User
		return t.updateUser(stub, args)
	} else if function == "deleteUser" {
		// Deletes an user by ID from the ledger
		return t.deleteUser(stub, args)
	} else if function == "getUsers" {
		// Get All Users from the ledger
		return t.getUsers(stub, args)
	}
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


func (t *SimpleChaincode) getUsers(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	startKey := args[0]
	endKey := args[1]
	limit, _ := strconv.Atoi(args[2])

	keysIter, err := stub.RangeQueryState(startKey, endKey)
	if err != nil {
		return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
	}
	defer keysIter.Close()

	var keys []string
	for keysIter.HasNext() {
		key, _, iterErr := keysIter.Next()
		if iterErr != nil {
			return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
		}
		keys = append(keys, key)
		limit = limit-1
		if limit <= 0 {
			break
		}
	}

	jsonKeys, err := json.Marshal(keys)
	if err != nil {
		return nil, fmt.Errorf("keys operation failed. Error marshaling JSON: %s", err)
	}

	return jsonKeys, nil
}



func (t *SimpleChaincode) updateUser(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	_,erval:=t. createOrUpdateUser(stub, args)
	return nil, erval
}


func (t *SimpleChaincode) deleteUser(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var userID string // user ID
	var err error
	var userIn User

	// validate input data for number of args, Unmarshaling to user state and obtain user id
	userIn, err = t.validateInput(args)
	if err != nil {
		return nil, err
	}
	userID = *userIn.UserID
	// Delete the key / user from the ledger
	err = stub.DelState(userID)
	if err != nil {
		err = errors.New("DELSTATE failed! : "+ fmt.Sprint(err))
		return nil, err
	}
	return nil, nil
}


func (t *SimpleChaincode) readUser(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var userID string
	var user User
	var err error

	stateIn, err:= t.validateInput(args)
	if err != nil {
		return nil, errors.New("User does not exist!")
	}
	userID = *stateIn.UserID

	// Get the user from the ledger
	userBytes, err:= stub.GetState(userID)
	if err != nil  || len(userBytes) == 0 {
		err = errors.New("Unable to get user state from ledger -- "+userID)
		return nil, err
	}
	err = json.Unmarshal(userBytes, &user)
	if err != nil {
		err = errors.New("Unable to unmarshal state data obtained from ledger -- "+userID)
		return nil, err
	}
	return userBytes, nil
}


func (t *SimpleChaincode) validateInput(args []string) (userIn User, err error) {
	var userID string
	var user User = User{}

	if len(args) !=1 {
		err = errors.New("Incorrect number of arguments. Expecting a JSON strings with mandatory userID")
		return user, err
	}
	jsonData := args[0]
	userID = ""
	stateJSON := []byte(jsonData)
	err = json.Unmarshal(stateJSON, &userIn)
	if err != nil {
		err = errors.New("Unable to unmarshal input JSON data")
		return user, err
	}

	if userIn.UserID != nil {
		userID = strings.TrimSpace(*userIn.UserID)
		if userID == "" {
			err = errors.New("UserID not passed")
			return user, err
		}
	} else {
		err = errors.New("UserID is mandatory in the input JSON data")
		return user, err
	}


	userIn.UserID = &userID
	return userIn, nil
}
//******************** createOrUpdateUser ********************/

func (t *SimpleChaincode) createOrUpdateUser(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var userID string
	var err error
	var userIn User
	var userStub User

	// validate input data for number of args, Unmarshaling to user state and obtain user id
	userIn, err = t.validateInput(args)
	if err != nil {
		return nil, err
	}
	userID = *userIn.UserID
	// Partial updates introduced here
	// Check if user record existed in stub
	userBytes, err := stub.GetState(userID)
	if err != nil || len(userBytes) == 0 { // Creat
		userStub = userIn
	} else { // Update
		err = json.Unmarshal(userBytes, &userStub)
		if err != nil {
			err = errors.New("Unable to unmarshal JSON data from stub")
			return nil, err
		}
		// Merge partial state updates
		userStub, err = t.mergePartialState(userStub, userIn)
		if err != nil {
			err = errors.New("Unable to merge state")
			return nil,err
		}

		/*if err := mergo.MergeWithOverwrite(&userStub, userIn); err != nil {
			err = errors.New("Unable to merge state")
			return nil,err
		}*/
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
