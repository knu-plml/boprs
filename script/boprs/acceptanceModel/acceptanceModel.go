package main
import (
    "fmt"
    "encoding/json"
    "errors"

    "github.com/hyperledger/fabric-chaincode-go/shim"
    sc "github.com/hyperledger/fabric-protos-go/peer"
)

var function string

type SmartContract struct {
}

func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) sc.Response {
    return shim.Success(nil)
}

func AccessControl(stub shim.ChaincodeStubInterface, function string, args []string) (UserType, string, error) {
    var userKey string
    var userType UserType
    var organization string
    var err error

    userType = REVIEWER
    userKey, err = GetIdentifier(stub)
    if err != nil {
        return userType, userKey, errors.New("AccessControl"+ " : " + err.Error())
    }

    switch function {
        case "QueryAcceptanceModel" :
            return userType, userKey, nil

        case "AddAcceptanceModel", "UpdateAcceptanceModel", "DeleteAcceptanceModel" :
            organization, err = GetMSPID(stub)
            if err != nil {
                return userType, userKey, errors.New("AccessControl" + " : " + err.Error())
            }

            err = CertifyEIC(stub, organization)
            if err != nil {
                return userType, userKey, errors.New("AccessControl" + " : " + err.Error())
            } else {
                userType = EIC
            }

            return userType, userKey, nil

        default :
            return userType, userKey, errors.New("AccessControl" + " : " + function + " is Invalid Smart Contract function name.")
    }
}

func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
    var args []string
    var err  error

    function, args := stub.GetFunctionAndParameters()

    _, _, err = AccessControl(stub, function, args)
    if err != nil {
        return shim.Error("acceptanceModel : " + err.Error())
    }

    switch function {
        case "QueryAcceptanceModel" :
            return s.QueryAcceptanceModel(stub, args)
        case "AddAcceptanceModel" :
            return s.AddAcceptanceModel(stub, args)
        case "UpdateAcceptanceModel" :
            return s.AddAcceptanceModel(stub, args)
        case "DeleteAcceptanceModel" :
            return s.DeleteAcceptanceModel(stub, args)
        default :
            return shim.Error(function + " is Invalid Smart Contract function name.")
    }
}

// args = [ organization ]
func (s *SmartContract) QueryAcceptanceModel(stub shim.ChaincodeStubInterface, args []string) sc.Response {
    var acceptanceModelByte []byte
    var err error

    if len(args) != 1 {
        return shim.Error(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    acceptanceModelByte, err = stub.GetState(args[0])
    if err != nil {
        return shim.Error(function + " : " + err.Error())
    }

    return shim.Success(acceptanceModelByte)
}

// args = [ OpenIdentity, OpenReport, OpenParticipation, OpenInteration, OpenPre-revieweManuscripts, OpenFinal-versionVommenting, OpenPaltform ]
func (s *SmartContract) AddAcceptanceModel(stub shim.ChaincodeStubInterface, args []string) sc.Response {
    var acceptanceModel AcceptanceModel
    var acceptanceModelByte []byte
    var organization string
    var err error

    if len(args) != 7 {
        return shim.Error(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    organization, err = GetMSPID(stub)
    if err != nil {
        return shim.Error(function + " : " + err.Error())
    }

    acceptanceModel = AcceptanceModel { Organization : organization, OI : args[0], OR : args[1], OP : args[2], ON : args[3], OM : args[4], OC : args[5], OPl : args[6] }

    acceptanceModelByte, err = json.Marshal(acceptanceModel)
    if err != nil {
        return shim.Error(function + " : " + err.Error())
    }

    err = stub.PutState(organization, acceptanceModelByte)
    if err != nil {
        return shim.Error(function + " : " + err.Error())
    }

    return shim.Success(nil)
}

// args = []
func (s *SmartContract) DeleteAcceptanceModel(stub shim.ChaincodeStubInterface, args []string) sc.Response {
    var organization string
    var err error

    if len(args) != 0 {
        return shim.Error(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    organization, err = GetMSPID(stub)
    if err != nil {
        return shim.Error(function + " : " + err.Error())
    }

    err = stub.DelState(organization)
    if err != nil {
        return shim.Error(function + " : " + err.Error())
    }

    return shim.Success(nil)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

    // Create a new Smart Contract
    err := shim.Start(new(SmartContract))
    if err != nil {
        fmt.Printf("Error creating new Smart Contract: %s", err)
    }
}
