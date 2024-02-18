package main
import (
    "fmt"
    "bytes"
    "encoding/json"
    "strconv"
    "time"
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

func AccessControl(stub shim.ChaincodeStubInterface, function string, args []string) (UserType, error) {
    var userType UserType
    var paperKey string
    var paper *Paper
    var err error

    userType = REVIEWER

    switch function {
        // editor of paper
        // args[0] = paperKey
        case "CreateContract" :
            paperKey = args[0]

            err = CheckPaperStatus(stub, args[0], "reviewer_invited")
            if err != nil {
                return userType, errors.New(function + " : " + err.Error())
            }

            err = CertifyEOPWithPaperKey(stub, paperKey)
            if err == nil {
                userType = EOP
            } else {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            return userType, nil
        // editor of paper
        // args[0] = contracteKey
        case "NextContract" :
            paperKey, err = CreatePaperKeyWithContractKey(stub, args[0])
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            paper, err = GetPaper(stub, paperKey)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            if paper.Status != "revise" {
                return userType, errors.New(function + " : " + paperKey + " paper is not in revise status.")
            }

            paperFileKey, err := CreatePaperFileKeyWithPaperKey(stub, paperKey, paper.Round + 1)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            _, err = InvokeChaincode(stub, "paper", "QueryPaperFile", []string{ paperFileKey })
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            err = CertifyEOPWithPaper(stub, paper)
            if err == nil {
                userType = EOP
            } else {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            return userType, nil
        // editor of paper
        // args[0] = contracteKey
        case "Fulfillment" :
            contractByte, err := stub.GetState(args[0])
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }
            if len(contractByte) == 0 {
                return userType ,errors.New("AccessControl" + " : " + args[0] + "contract does not exist.")
            }

            paperKey, err = CreatePaperKeyWithContractKey(stub, args[0])
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            if CheckPaperStatus(stub, paperKey, "reviewer_invited") != nil && CheckPaperStatus(stub, paperKey, "revise") != nil {
                return userType, errors.New(function + " : " + "Review status is not \"reviewer_invited\" or \"revise\".")
            }

            err = CertifyEOPWithPaperKey(stub, paperKey)
            if err == nil {
                userType = EOP
            } else {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            return userType, nil
        // editor of paper
        // args[0] = contracteKey
        case "CompleteContract" :
            paperKey, err = CreatePaperKeyWithContractKey(stub, args[0])
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            err = CheckPaperStatus(stub, paperKey, "under_review")
            if err != nil {
                return userType, errors.New(function + " : " + err.Error())
            }

            err = CertifyEOPWithPaperKey(stub, paperKey)
            if err == nil {
                userType = EOP
            } else {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            return userType, nil
        // editor of paper, editor of organization
        // args[0] = contracteKey
        case "QuerySignatureWithContractKey" :
            paperKey, err = CreatePaperKeyWithContractKey(stub, args[0])
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            paper, err = GetPaper(stub, paperKey)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            err = CertifyEOPWithPaper(stub, paper)
            if err == nil {
                userType = EOP
            } else {
                err = CertifyEOOWithPaper(stub, paper)
                if err == nil {
                    userType = EOO
                } else {
                    return userType, errors.New("AccessControl"+ " : " + err.Error())
                }
            }

            return userType, nil
        // editor of paper or editor of organization or author or reviewer of paper
        // args[0] = contractKey
        case "QueryContract" :
            paperKey, err = CreatePaperKeyWithContractKey(stub, args[0])
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            paper, err = GetPaper(stub, paperKey)
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            err = CertifyEOPWithPaper(stub, paper)
            if err == nil {
                userType = EOP
            } else {
                err = CertifyEOOWithPaper(stub, paper)
                if err == nil {
                    userType = EOO
                } else {
                    err = CertifyAuthorWithPaper(stub, paper)
                    if err == nil {
                        userType = AUTHOR
                    } else {
                        err = CertifyROPWithPaperKey(stub, paperKey)
                        if err == nil {
                            userType = ROP
                        } else {
                            return userType, errors.New("AccessControl"+ " : " + err.Error())
                        }
                    }
                }
            }

            return userType, nil
        // editor of paper or editor of organization or author or reviewer of paper
        // args[0] = paperKey
        case "QueryContractWithPaperKey" :
            paperKey = args[0]
            paper, err = GetPaper(stub, paperKey)
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            err = CertifyEOPWithPaper(stub, paper)
            if err == nil {
                userType = EOP
            } else {
                err = CertifyEOOWithPaper(stub, paper)
                if err == nil {
                    userType = EOO
                } else {
                    err = CertifyAuthorWithPaper(stub, paper)
                    if err == nil {
                        userType = AUTHOR
                    } else {
                        err = CertifyROPWithPaperKey(stub, paperKey)
                        if err == nil {
                            userType = ROP
                        } else {
                            return userType, errors.New("AccessControl"+ " : " + err.Error())
                        }
                    }
                }
            }

            return userType, nil
        // reviewer of paper
        // args[0] = contractKey
        case "SignContract" :
            paperKey, err = CreatePaperKeyWithContractKey(stub, args[0])
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            err = CheckContractStatus(stub, args[0], "reviewer_invited")
            if err != nil {
                return userType, errors.New(function + " : " + err.Error())
            }

            err = CertifyROPWithPaperKey(stub, paperKey)
            if err == nil {
                userType = ROP
            } else {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            return userType, nil
        // reviewer of paper himself or editor of paper or editor of organization
        case "QuerySignature" :
            paperKey, err = CreatePaperKeyWithContractKey(stub, args[0])
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            paper, err = GetPaper(stub, paperKey)
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            err = CertifyEOPWithPaper(stub, paper)
            if err == nil {
                userType = EOP
            } else {
                err = CertifyEOOWithPaper(stub, paper)
                if err == nil {
                    userType = EOO
                } else {
                    _, splitedSigKey, err := stub.SplitCompositeKey(args[0])
                    if err != nil {
                        return userType, errors.New("AccessControl" + " : " + err.Error())
                    }

                    err = CertifyIdentifier(stub, splitedSigKey[3])
                    if err == nil {
                        userType = SELF
                    } else {
                        return userType, errors.New("AccessControl"+ " : " + err.Error())
                    }
                }
            }

            return userType, nil

        // editor of paper or author
        // args[0] = paperKey
        case "DeleteContractWithPaperKey" :
            paperKey = args[0]

            err = CertifyEOPWithPaperKey(stub, paperKey)
            if err == nil {
                userType = EOP
            } else {
                err = CertifyAuthorWithPaperKey(stub, paperKey)
                if err == nil {
                    userType = AUTHOR
                } else {
                    return userType, errors.New("AccessControl"+ " : " + err.Error())
                }
            }

            return userType, nil

        default :
            return userType, errors.New("AccessControl"+ " : " + function + " is Invalid Smart Contract function name.")
    }
}

func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
    var args []string
    var result []byte
    var err error

    //get function name and argument list
    function, args = stub.GetFunctionAndParameters()

    //access control
    _, err = AccessControl(stub, function, args)
    if err != nil {
        return shim.Error(err.Error())
    }

    switch function {
        // No return
        case "CreateContract" :
            err = s.CreateContract(stub, args)
        case "NextContract" :
            err = s.NextContract(stub, args)
        case "Fulfillment" :
            err = s.Fulfillment(stub, args)
        case "CompleteContract" :
            err = s.CompleteContract(stub, args)
        case "SignContract" :
            err = s.SignContract(stub, args)
        case "DeleteContractWithPaperKey" :
            err = s.DeleteContractWithPaperKey(stub, args)

        // return Contract
        case "QueryContract" :
            result, err = s.QueryContract(stub, args)

        // return Contract list
        case "QueryContractWithPaperKey" :
            result, err = s.QueryContractWithPaperKey(stub, args)

        // return Signature
        case "QuerySignature" :
            result, err = s.QuerySignature(stub, args)

        // return Signature list
        case "QuerySignatureWithContractKey" :
            result, err = s.QuerySignatureWithContractKey(stub, args)
    }

    if err != nil {
        return shim.Error(err.Error())
    }

    return shim.Success(result)
}

//args = [ paperKey, dueDate ]
func (s *SmartContract) CreateContract(stub shim.ChaincodeStubInterface, args []string) error {
    var contract Contract
    var contractByte []byte
    var paper *Paper
    var round int
    var roundString string
    var key string
    var err error

    if len(args) != 2 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 2")
    }

    paper, err = GetPaper(stub, args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if paper.Round != 0 {
        return errors.New(function + " : " + "Call \"createContract\" to write the first contract")
    }

    round = 1
    roundString = "1"

    key, err = CreateContractKeyWithPaperKey(stub, args[0], roundString)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    contract = Contract { Key : key, PaperKey : args[0], Round : round, DueDate : args[1], CompleteDate : "" }

    contractByte, err = json.Marshal(contract)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(key, contractByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    _, err = InvokeChaincode(stub, "paper", "UpdateRound", []string { args[0], roundString })
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    _, err = InvokeChaincode(stub, "message", "InitMessage", []string{ key })
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

//args = [ contractKey, dueDate ]
func (s *SmartContract) NextContract(stub shim.ChaincodeStubInterface, args []string) error {
    var contract Contract
    var contractByte []byte
    var paperKey string
    var paper *Paper
    var round int
    var roundString string
    var key string
    var err error

    if len(args) != 2 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 2")
    }

    contractByte, err = stub.GetState(args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if len(contractByte) == 0 {
        return errors.New(function + " : " + args[0] + "contract does not exist.")
    }

    paperKey, err = CreatePaperKeyWithContractKey(stub, args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    paper, err = GetPaper(stub, paperKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if paper.Round == 0 {
        return errors.New(function + " : " + "Call \"createContract\" to write the first contract")
    }

    round = paper.Round + 1
    roundString = strconv.Itoa(round)

    key, err = CreateContractKeyWithPaperKey(stub, paperKey, roundString)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    contract = Contract { Key : key, PaperKey : paperKey, Round : round, DueDate : args[1], CompleteDate : "" }

    contractByte, err = json.Marshal(contract)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(key, contractByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    _, err = InvokeChaincode(stub, "reviewer", "InitReviewer", []string { paperKey })
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    _, err = InvokeChaincode(stub, "paper", "UpdateRound", []string { paperKey, roundString })
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    _, err = InvokeChaincode(stub, "message", "InitMessage", []string{ key })
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

//args = [ contractKey ]
func (s *SmartContract) QueryContract(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var contractByte []byte
    var err error

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    contractByte, err = stub.GetState(args[0])
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if len(contractByte) == 0 {
        return nil, errors.New(function + " : " + "The " + args[0] + " contract could not be found.")
    }

    return contractByte, nil
}

//args = [ paperKey ]
func (s *SmartContract) QueryContractWithPaperKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var contractIter shim.StateQueryIteratorInterface
    var splitPaperKey []string
    var buff bytes.Buffer
    var flag bool
    var err error

    flag = false

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    _, splitPaperKey, err = stub.SplitCompositeKey(args[0])
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }

    contractIter, err = stub.GetStateByPartialCompositeKey("", splitPaperKey)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !contractIter.HasNext() {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for contractIter.HasNext() {
        contract, err := contractIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if flag {
            buff.WriteString(",")
        } else {
            flag = true
        }

        buff.Write(contract.Value)
    }
    buff.WriteString("]")

    return buff.Bytes(), nil
}

//args = [ contractKey ]
func (s *SmartContract) Fulfillment(stub shim.ChaincodeStubInterface, args []string) error {
    var paperKey string
    var reviewersByte []byte
    var reviewersArray []Reviewer
    var err error

    if len(args) != 1 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    paperKey, err = CreatePaperKeyWithContractKey(stub, args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    reviewersByte, err = InvokeChaincode(stub, "reviewer", "QueryReviewerWithPaperKey", []string { paperKey })
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = json.Unmarshal(reviewersByte, &reviewersArray)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    _, err = InvokeChaincode(stub, "paper", "UpdateStatus", []string { paperKey, "under_review" })
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

//args = [ contractKey ]
func (s *SmartContract) CompleteContract (stub shim.ChaincodeStubInterface, args []string) error {
    var contract Contract
    var contractByte []byte
    var paperKey string
    var err error

    if len(args) != 1 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    paperKey, err = CreatePaperKeyWithContractKey(stub, args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    completeTime, err := stub.GetTxTimestamp()
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    contractByte, err = stub.GetState(args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if len(contractByte) == 0 {
        return errors.New(function + " : " + args[0] + " contract is not exist.")
    }

    err = json.Unmarshal(contractByte, &contract)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    contract.CompleteDate = time.Unix(completeTime.GetSeconds(), 0).String()

    contractByte, err = json.Marshal(contract)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(args[0], contractByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    _, err = InvokeChaincode(stub, "paper", "UpdateStatus", []string { paperKey, "under_decision"})
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

//args = [ contractKey, signature ]
func (s *SmartContract) SignContract(stub shim.ChaincodeStubInterface, args []string) error {
    var signatureKey string
    var paperKey string
    var identifier string
    var PRReviewerKeyByte []byte
    var PRReviewerKey string
    var signature Signature
    var signatureByte []byte
    var err error

    if len(args) != 2 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 2")
    }

    paperKey, err = CreatePaperKeyWithContractKey(stub, args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    identifier, err = GetIdentifier(stub)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    PRReviewerKeyByte, err = InvokeChaincode(stub, "reviewer", "GetPRKeyWithPaperKey", []string { paperKey, identifier })
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    PRReviewerKey = string(PRReviewerKeyByte)

    err = CheckReviewerStatus(stub, PRReviewerKey, "selected")
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    signatureKey, err = CreateSignatureKeyWithContractKey(stub, args[0], identifier)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    signature = Signature { Key : signatureKey, ContractKey : args[0], PaperKey : paperKey, ReviewerKey: PRReviewerKey, Signature : args[1] }
    signatureByte, err = json.Marshal(signature)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(signatureKey, signatureByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    _, err = InvokeChaincode(stub, "reviewer", "UpdateStatus", []string { PRReviewerKey, "accept" })
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

//args = [ paperKey ]
func (s *SmartContract) DeleteContractWithPaperKey(stub shim.ChaincodeStubInterface, args []string) error {
    var contractIter shim.StateQueryIteratorInterface
    var splitPaperKey []string
    var err error

    if len(args) != 1 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    _, splitPaperKey, err = stub.SplitCompositeKey(args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    contractIter, err = stub.GetStateByPartialCompositeKey("", splitPaperKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if !contractIter.HasNext() {
        return nil
    }

    for contractIter.HasNext() {
        contractKV, err := contractIter.Next()
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        err = stub.DelState(contractKV.Key)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }
    }

    return nil
}

//args = [ signatureKey ]
func (s *SmartContract) QuerySignature(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var signatureByte []byte
    var err error

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    signatureByte, err = stub.GetState(args[0])
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if len(signatureByte) == 0 {
        return nil, errors.New(function + " : " + args[0] + " signature does not exists.")
    }

    return signatureByte, nil
}

//args = [ contractKey ]
func (s *SmartContract) QuerySignatureWithContractKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var splitContractKey []string
    var signatureIter shim.StateQueryIteratorInterface
    var buff bytes.Buffer
    var flag bool
    var err error

    flag = false

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    _, splitContractKey, err = stub.SplitCompositeKey(args[0])
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }

    signatureIter, err = stub.GetStateByPartialCompositeKey("SIG", splitContractKey)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !signatureIter.HasNext() {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for signatureIter.HasNext() {
        signatureKV, err := signatureIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if flag {
            buff.WriteString(", ")
        } else {
            flag = true
        }

        buff.Write(signatureKV.Value)
    }

    buff.WriteString("]")
    signatureIter.Close()

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

