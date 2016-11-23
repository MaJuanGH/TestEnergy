package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type SimpleChaincode struct {
}

var homeNo int = 0
var transactionNo int = 0

type Home struct {
	Address string
	Energy  int
	Money   int
	Id      int
	Status  int
	PriKey  string
	PubKey  string
}

type Transaction struct {
	BuyerAddress     string
	BuyerAddressSign string
	SellerAddress    string
	Energy           int
	Money            int
	Id               int
	Time             int64
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	fmt.Printf("Init OK!\n")
	return nil, nil
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function == "changeStatus" {
		if len(args) != 3 {
			return nil, errors.New("Incorrect number of arguments. Expecting 1")
		}
		return changeStatus(stub, args)
	} else if function == "buyByAddress" {
		if len(args) != 4 {
			return nil, errors.New("Incorrect number of arguments. Expecting 1")
		}
		return buyByAddress(stub, args)
	} else if function == "createUser" {
		if len(args) != 2 {
			return nil, errors.New("Incorrect number of arguments. Expecting 1")
		}
		return t.createUser(stub, args)
	}
	return nil, errors.New("Received unknown function invocation")
}

func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function == "getHomeByAddress" {
		if len(args) != 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting 1")
		}
		_, homeBytes, err := getHomeByAddress(stub, args[0])
		if err != nil {
			fmt.Println("Error get home\n")
			return nil, err
		}
		return homeBytes, nil
	} else if function == "getHomes" {
		if len(args) != 0 {
			return nil, errors.New("Incorrect number of arguments. Expecting 0")
		}
		homes, err := getHomes(stub)
		if err != nil {
			fmt.Println("Error unmarshalling")
			return nil, err
		}
		homeBytes, err1 := json.Marshal(&homes)
		if err1 != nil {
			fmt.Println("Error marshalling banks\n")
		}
		return homeBytes, nil
	} else if function == "getTransactions" {
		if len(args) != 0 {
			return nil, errors.New("Incorrect number of arguments. Expecting 0")
		}
		transactions, err := getTransactions(stub)
		if err != nil {
			fmt.Println("Error unmarshalling\n")
			return nil, err
		}
		txBytes, err1 := json.Marshal(&transactions)
		if err1 != nil {
			fmt.Println("Error marshalling data\n")
		}
		return txBytes, nil
	} else if function == "getTransactionById" {
		if len(args) != 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting 1")
		}
		_, txBytes, err := getTransactionById(stub, args[0])
		if err != nil {
			return nil, err
		}
		return txBytes, nil
	}
	return nil, errors.New("Received unknown function invocation")
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s\n", err)
	}
}

//生成Address
func GetAddress() (string, string, string) {
	var address, priKey, pubKey string
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		fmt.Printf("get rand failed\n")
		return "", "", ""
	}

	h := md5.New()
	h.Write([]byte(base64.URLEncoding.EncodeToString(b)))

	address = hex.EncodeToString(h.Sum(nil))
	priKey = address + "1"
	pubKey = address + "2"

	return address, priKey, pubKey
}

func (t *SimpleChaincode) createUser(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("Enter createUser...\n")
	var energy, money int
	var err error
	var homeBytes []byte
	if len(args) != 2 {
		fmt.Printf("args != 2\n")
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
	address, priKey, pubKey := GetAddress()
	energy, err = strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("get energy failed!\n")
		return nil, errors.New("want Integer number")
	}
	money, err = strconv.Atoi(args[1])
	if err != nil {
		fmt.Printf("get money failed\n")
		return nil, errors.New("want Integer number")
	}
	fmt.Printf("HomeInfo: address = %v, energy = %v, money = %v, homeNo = %v, priKey = %v, pubKey = %v\n", address, energy, money, homeNo, priKey, pubKey)
	home := Home{Address: address, Energy: energy, Money: money, Id: homeNo, Status: 1, PriKey: priKey, PubKey: pubKey}
	err = writeHome(stub, home)
	if err != nil {
		fmt.Printf("writehome failed\n")
		return nil, errors.New("write Error" + err.Error())
	}
	homeBytes, err = json.Marshal(&home)
	if err != nil {
		fmt.Printf("marshal home byte failed\n")
		return nil, errors.New("Error retrieve")
	}
	homeNo = homeNo + 1
	fmt.Printf("Create user success!\n")
	return homeBytes, nil
}

func buyByAddress(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("Enter buyByAddress\n")
	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}
	homeSeller, _, err := getHomeByAddress(stub, args[0])
	homeBuyer, _, err := getHomeByAddress(stub, args[2])

	if args[1] != args[2]+"11" {
		fmt.Printf("Verify sign data failed!\n")
		return nil, errors.New("Verify sign data failed!")
	}

	fmt.Printf("Verify sign data ok\n")
	buyValue, erro := strconv.Atoi(args[3])
	if erro != nil {
		fmt.Printf("The last args should be a integer number\n")
		return nil, errors.New("want integer number")
	}
	if homeSeller.Energy < buyValue && homeBuyer.Money < buyValue {
		fmt.Printf("not enough money or energy\n")
		return nil, errors.New("not enough money or energy")
	}

	fmt.Printf("Before trans:\n  homeSeller.Energy = %d, homeSeller.Money = %d\n", homeSeller.Energy, homeSeller.Money)
	fmt.Printf("  homeBuyer.Energy = %d, homeBuyer.Money = %d\n", homeBuyer.Energy, homeBuyer.Money)

	homeSeller.Energy = homeSeller.Energy - buyValue
	homeSeller.Money = homeSeller.Money + buyValue
	homeBuyer.Energy = homeBuyer.Energy + buyValue
	homeBuyer.Money = homeBuyer.Money - buyValue

	fmt.Printf("After trans:\n  homeSeller.Energy = %d, homeSeller.Money = %d\n", homeSeller.Energy, homeSeller.Money)
	fmt.Printf("  homeBuyer.Energy = %d, homeBuyer.Money = %d\n", homeBuyer.Energy, homeBuyer.Money)

	err = writeHome(stub, homeSeller)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Write homeSeller OK!\n")

	err = writeHome(stub, homeBuyer)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Write homeBuyer OK!\n")

	fmt.Printf("TransactionInfo:\n")
	fmt.Printf("    BuyerAddress: %v\n", args[2])
	fmt.Printf("    BuyerAddressSign: %v\n", args[1])
	fmt.Printf("    SellerAddress: %v\n", args[0])
	fmt.Printf("    Energy: %v\n", buyValue)
	fmt.Printf("    Money: %v\n", buyValue)
	fmt.Printf("    Id: %v\n", transactionNo)

	transaction := Transaction{BuyerAddress: args[2], BuyerAddressSign: args[1], SellerAddress: args[0], Energy: buyValue, Money: buyValue, Id: transactionNo, Time: time.Now().Unix()}
	err = writeTransaction(stub, transaction)
	if err != nil {
		return nil, err
	}
	transactionNo = transactionNo + 1
	txBytes, err := json.Marshal(&transaction)

	if err != nil {
		return nil, errors.New("Error retrieving schoolBytes")
	}

	return txBytes, nil
}

func changeStatus(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}
	home, homeBytes, err := getHomeByAddress(stub, args[0])
	if err != nil {
		return nil, err
	}

	if args[1] == args[0]+"11" {
		status, _ := strconv.Atoi(args[2])
		home.Status = status
		err = writeHome(stub, home)
		if err != nil {
			return homeBytes, nil
		}
	}
	return nil, err
}

func getHomeByAddress(stub shim.ChaincodeStubInterface, address string) (Home, []byte, error) {
	var home Home
	homeBytes, err := stub.GetState(address)
	if err != nil {
		fmt.Println("Error retrieving home")
	}
	err = json.Unmarshal(homeBytes, &home)
	if err != nil {
		fmt.Println("Error unmarshalling home")
	}
	return home, homeBytes, nil
}

func getHomes(stub shim.ChaincodeStubInterface) ([]Home, error) {
	var homes []Home
	var number string
	var err error
	var home Home
	if homeNo <= 10 {
		i := 0
		for i < homeNo {
			number = strconv.Itoa(i)
			home, _, err = getHomeByAddress(stub, number)
			if err != nil {
				return nil, errors.New("Error get detail")
			}
			homes = append(homes, home)
			i = i + 1
		}
	} else {
		i := 0
		for i < 10 {
			number = strconv.Itoa(i)
			home, _, err = getHomeByAddress(stub, number)
			if err != nil {
				return nil, errors.New("Error get detail")
			}
			homes = append(homes, home)
			i = i + 1
		}
		return homes, nil
	}
	return nil, nil
}

func getTransactionById(stub shim.ChaincodeStubInterface, id string) (Transaction, []byte, error) {
	var transaction Transaction
	txBytes, err := stub.GetState("transaction" + id)
	if err != nil {
		fmt.Println("Error retrieving home")
	}

	err = json.Unmarshal(txBytes, &transaction)
	if err != nil {
		fmt.Println("Error unmarshalling home")
	}

	return transaction, txBytes, nil
}

func getTransactions(stub shim.ChaincodeStubInterface) ([]Transaction, error) {
	var transactions []Transaction
	var number string
	var err error
	var transaction Transaction
	if transactionNo <= 10 {
		i := 0
		for i < transactionNo {
			number = strconv.Itoa(i)
			transaction, _, err = getTransactionById(stub, number)
			if err != nil {
				return nil, errors.New("Error get detail")
			}
			transactions = append(transactions, transaction)
			i = i + 1
		}
		return transactions, nil
	} else {
		i := 0
		for i < 10 {
			number = strconv.Itoa(i)
			transaction, _, err = getTransactionById(stub, number)
			if err != nil {
				return nil, errors.New("Error get detail")
			}
			transactions = append(transactions, transaction)
			i = i + 1
		}
		return transactions, nil
	}
	return nil, nil
}

func writeHome(stub shim.ChaincodeStubInterface, home Home) error {
	fmt.Printf("Enter writeHome \n")
	homeBytes, err := json.Marshal(&home)
	if err != nil {
		fmt.Printf("json.Marshal failed \n")
		return err
	}
	err = stub.PutState(home.Address, homeBytes)
	if err != nil {
		fmt.Printf("stub.PutState failed \n")
		return errors.New("PutState Error" + err.Error())
	}
	fmt.Printf("Out writeHome \n")
	return nil
}

func writeTransaction(stub shim.ChaincodeStubInterface, transaction Transaction) error {
	fmt.Printf("Enter writeTransaction \n")
	txBytes, err := json.Marshal(&transaction)
	if err != nil {
		fmt.Printf("json.Marshal failed \n")
		return nil
	}

	id := strconv.Itoa(transaction.Id)
	err = stub.PutState("transaction"+id, txBytes)
	if err != nil {
		fmt.Printf("stub.PutState failed \n")
		return errors.New("PutState Error" + err.Error())
	}
	fmt.Printf("Out writeTransaction \n")
	return nil
}
