package main
import (
    "encoding/json"
    "fmt"
    "bytes"
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

//func AccessControl(stub shim.ChaincodeStubInterface, function string, args []string) (UserType, error) {
//    var userType UserType
//    var paperKey string
//    var paper *Paper
//    var acceptanceModel *AcceptanceModel
//    var err error
//
//    userType = REVIEWER
//
//    switch function {
//        //editor of paper
//        // args[0] = contractKey
//        case "AddReport" :
//            paperKey, err = CreatePaperKeyWithContractKey(stub, args[0])
//            if err != nil {
//                return userType, errors.New("AccessControl"+ " : " + err.Error())
//            }
//
//            err = CheckContractStatus(stub, args[0], "under_decision")
//            if err != nil {
//                return userType, errors.New("AccessControl"+ " : " + err.Error())
//            }
//
//            err = CertifyEOPWithPaperKey(stub, paperKey)
//            if err == nil {
//                userType = EOP
//            } else {
//                return userType, errors.New("AccessControl"+ " : " + err.Error())
//            }
//
//            return userType, nil
//
//        // broker
//        // args[0] = contractKey
//        case "AddReportWithEditorID" :
//            err = CertifyBroker(stub)
//            if err != nil {
//                return userType, errors.New("AccessControl"+ " : " + err.Error())
//            } else {
//                userType = EOO
//            }
//
//            return userType, nil
//
//        //editor of paper or author or reviewer of paper
//        //args[0] = reportKey
//        case "QueryReport" :
//            paperKey, err = CreatePaperKeyWithContractKey(stub, args[0])
//            if err != nil {
//                return userType, errors.New("AccessControl" + " : " + err.Error())
//            }
//
//            err = CertifyEOPWithPaperKey(stub, paperKey)
//            if err == nil {
//                userType = EOP
//            } else {
//                err = CertifyAuthorWithPaperKey(stub, paperKey)
//                if err == nil {
//                    userType = AUTHOR
//                } else {
//                    err = CertifyROPWithPaperKey(stub, paperKey)
//                    if err == nil {
//                        userType = ROP
//                    } else {
//                        paper, err = GetPaper(stub, paperKey)
//                        if err != nil {
//                            return userType, errors.New("AccessControl" + " : " + err.Error())
//                        }
//
//                        acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
//                        if err != nil {
//                            return userType, errors.New("AccessControl" + " : " + err.Error())
//                        }
//
//                        if !ORCheck(paper, acceptanceModel) {
//                            return userType, errors.New("AccessControl" + " : " + "OR error")
//                        }
//                    }
//                }
//            }
//
//            return userType, nil
//        //editor of paper or author or reviewer of paper
//        //args[0] = paperKey
//        case "QueryReportWithPaperKey" :
//            paperKey = args[0]
//
//            err = CertifyEOPWithPaperKey(stub, paperKey)
//            if err == nil {
//                userType = EOP
//            } else {
//                err = CertifyAuthorWithPaperKey(stub, paperKey)
//                if err == nil {
//                    userType = AUTHOR
//                } else {
//                    err = CertifyROPWithPaperKey(stub, paperKey)
//                    if err == nil {
//                        userType = ROP
//                    } else {
//                        paper, err = GetPaper(stub, paperKey)
//                        if err != nil {
//                            return userType, errors.New("AccessControl" + " : " + err.Error())
//                        }
//
//                        acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
//                        if err != nil {
//                            return userType, errors.New("AccessControl" + " : " + err.Error())
//                        }
//
//                        if !ORCheck(paper, acceptanceModel) {
//                            return userType, errors.New("AccessControl" + " : " + "OR error")
//                        }
//                    }
//                }
//            }
//
//            return userType, nil
//        default :
//            return userType, errors.New("AccessControl"+ " : " + function + " is Invalid Smart Contract function name.")
//        }
//}

func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
    var args []string
    var result []byte = nil
    var err error

    //get function name and argument list
    function, args = stub.GetFunctionAndParameters()

    //access control
//    _, err = AccessControl(stub, function, args)
//    if err != nil {
//        return shim.Error(err.Error())
//    }

    switch function {
        // No Return
        case "AddReport" :
            err = s.AddReport(stub, args)

        case "AddReportWithEditorID" :
            err = s.AddReportWithEditorID(stub, args)

        // Report
        case "QueryReport" :
            result, err = s.QueryReport(stub, args)

        // Report Array
        case "QueryReportWithPaperKey" :
            result, err = s.QueryReportWithPaperKey(stub, args)
    }

    if err != nil {
        return shim.Error(err.Error())
    }

    return shim.Success(result)
}

//args = [ contractKey, location, comment, decision ]
func (s *SmartContract) AddReport(stub shim.ChaincodeStubInterface, args []string) error {
    var reportByte []byte
    var report Report
    var identifier string
    var overall Comment
    var OAlocation []string
    var OAcomment []string
    var paperKey string
    var err error

    if len(args) != 4 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 4")
    }

    paperKey, err = CreatePaperKeyWithContractKey(stub, args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    identifier, err = GetIdentifier(stub)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = json.Unmarshal([]byte(args[1]), &OAlocation)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = json.Unmarshal([]byte(args[2]), &OAcomment)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    overall = Comment { ContractKey : args[0], ReviewerID : identifier, Location : OAlocation, Comment : OAcomment }
    report = Report { Key : args[0], ContractKey : args[0], OverallComment : overall, Decision : args[3] }
    reportByte, err = json.Marshal(report)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(args[0], reportByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    _, err = InvokeChaincode(stub, "test_paper", "UpdateStatus", []string { paperKey, args[3] })
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

//args = [ contractKey, location, comment, decision, editorID ]
func (s *SmartContract) AddReportWithEditorID(stub shim.ChaincodeStubInterface, args []string) error {
    var reportByte []byte
    var report Report
    var identifier string
    var overall Comment
    var OAlocation []string
    var OAcomment []string
    var paperKey string
    var err error

    if len(args) != 5 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 5")
    }

    paperKey, err = CreatePaperKeyWithContractKey(stub, args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = json.Unmarshal([]byte(args[1]), &OAlocation)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = json.Unmarshal([]byte(args[2]), &OAcomment)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    identifier = args[4]

    overall = Comment { ContractKey : args[0], ReviewerID : identifier, Location : OAlocation, Comment : OAcomment }
    report = Report { Key : args[0], ContractKey : args[0], OverallComment : overall, Decision : args[3] }
    reportByte, err = json.Marshal(report)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(args[0], reportByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    _, err = InvokeChaincode(stub, "test_paper", "UpdateStatus", []string { paperKey, args[3] })
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

//args = [ reportKey ]
func (s *SmartContract) QueryReport(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var reportByte []byte
    var err error

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    reportByte, err = stub.GetState(args[0])
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if len(reportByte) == 0 {
        return nil, errors.New(function + " : " + "There are no report for " + args[0] + " contract.")
    }

    return reportByte, nil
}

//args = [ paperKey ]
func (s *SmartContract) QueryReportWithPaperKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var reportIter shim.StateQueryIteratorInterface
    var splitPaperKey []string
    var buff bytes.Buffer
    var err error

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    _, splitPaperKey, err = stub.SplitCompositeKey(args[0])
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }

    reportIter, err = stub.GetStateByPartialCompositeKey("", splitPaperKey)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !reportIter.HasNext() {
        return []byte("[]"), nil
    }

    reportKV, err := reportIter.Next()
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }

    buff.WriteString("[")
    buff.Write(reportKV.Value)

    for reportIter.HasNext() {
        reportKV, err = reportIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        buff.WriteString(",")
        buff.Write(reportKV.Value)
    }

    buff.WriteString("]")
    reportIter.Close()

    return buff.Bytes(), nil
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

    // Create a new Smart Contract
    err := shim.Start(new(SmartContract))
    if err != nil {
        fmt.Printf("Error creating new Smart Contract: %s", err)
    }
}
