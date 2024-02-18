package main
import (
    "encoding/json"
    "fmt"
    "bytes"
    "errors"
    "strconv"
    "time"

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
    var acceptanceModel *AcceptanceModel
    var err error

    userType = REVIEWER
    if err != nil {
        return userType, errors.New("AccessControl"+ " : " + err.Error())
    }

    switch function {
        // AUTHOR, EOP, ROP
        // ON, OI
        // args[0] = parent key
        case "AddMessage" :
            paperKey, err := CreatePaperKeyWithMessageKey(stub, args[0])
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            userType, err = GetUserTypeWithPaperKey(stub, paperKey)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            err = CheckReviewEndWithPaperKey(stub, paperKey)
            if err != nil {
                if userType != ROP && userType != AUTHOR && userType != EOP {
                    return userType, errors.New("AccessControl" + " : " + "AccessError")
                } else {
                    return userType, nil
                }
            } else {
                return userType, nil
            }

        // EOP
        // args[0] = contractKey
        case "InitMessage" :
            paperKey, err := CreatePaperKeyWithMessageKey(stub, args[0])
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            userType, err = GetUserTypeWithPaperKey(stub, paperKey)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            if userType == EOP {
                return userType, nil
            } else {
                return userType, errors.New("AccessControl" + " : " + "AccessError")
            }

        // AUTHOR, EOP, ROP
        // after review
        case "QueryMessage" :
            paperKey, err := CreatePaperKeyWithMessageKey(stub, args[0])
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            userType, err = GetUserTypeWithPaperKey(stub, paperKey)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            if userType == ROP || userType == AUTHOR || userType == EOP {
                return userType, nil
            }

            paper, err := GetPaper(stub, paperKey)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            if ONCheck(paper, acceptanceModel) {
                return userType, nil
            }

            return userType, errors.New("AccessControl" + " : " + "ON error")

        case "QueryMessageWithPaperKey" :
            paperKey := args[0]
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            userType, err = GetUserTypeWithPaperKey(stub, paperKey)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            if userType == ROP || userType == AUTHOR || userType == EOP {
                return userType, nil
            }

            paper, err := GetPaper(stub, paperKey)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            if ONCheck(paper, acceptanceModel) {
                return userType, nil
            } else {
                return AERROR, nil
            }

        case "QueryMessageWithContractKey" :
            paperKey, err := CreatePaperKeyWithContractKey(stub, args[0])
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            userType, err = GetUserTypeWithPaperKey(stub, paperKey)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            if userType == ROP || userType == AUTHOR || userType == EOP {
                return userType, nil
            }

            paper, err := GetPaper(stub, paperKey)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            roundString, err := GetRoundFromContractKey(stub, args[0])
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            targetRound, err := strconv.Atoi(roundString)
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            if targetRound > paper.Round {
                return userType, errors.New("AccessControl" + " : " + "The comment does not exist.")
            }

            acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            if (targetRound < paper.Round && (acceptanceModel.ON == "4" || acceptanceModel.ON == "5")) || ONCheck(paper, acceptanceModel) {
                userType = REVIEWER
                return userType, nil
            } else {
                return AERROR, nil
            }

        // SELF
        // OI
        case "QueryMessageWithReviewerID" :
            err = CertifyIdentifier(stub, args[0])
            if err == nil {
                userType = SELF
                return userType, nil
            } else {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

        // editor of paper or author
        // args[0] = paperKey
        case "DeleteMessageWithPaperKey" :
            paperKey := args[0]

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
            return userType, errors.New("AccessControl" + " : " + "function name error(" + function + ")")
    }
}

func ApplyAcceptanceModel(stub shim.ChaincodeStubInterface, function string, userType UserType, target []byte) ([]byte, error) {
    var paper *Paper
    var message Message
    var messageByte []byte
    var messageArray []Message
    var acceptanceModel *AcceptanceModel
    var buff bytes.Buffer
    var writeOne bool
    var identifier string
    var err error

    writeOne = false

    switch function {
        // No Return
        // OI
        case "AddMessage", "InitMessage" :
            return nil, errors.New("ApplyAcceptanceModel" + " : " + "function errer(" + function + ")")

        // Message
        // OI
        case "QueryMessage" :
            if len(target) == 0 {
                return nil, errors.New("ApplyAcceptanceModel" + " : " + "The comment does not exist.")
            }

            identifier, err = GetIdentifier(stub)
            if err != nil {
                return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
            }

            if userType == EOP {
                return target, nil
            } else if userType == ROP {
                err = json.Unmarshal(target, &message)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                if message.ReviewerID == identifier {
                    return target, nil
                } else {
                    paper, err = GetPaper(stub, message.PaperKey)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    if !OICheck(paper, acceptanceModel) {
                        message.ReviewerID = ""
                        target, err = json.Marshal(paper)
                    }

                    return target, nil
                }

            } else {
                err = json.Unmarshal(target, &message)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                paper, err = GetPaper(stub, message.PaperKey)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                if !OICheck(paper, acceptanceModel) {
                    message.ReviewerID = ""
                    target, err = json.Marshal(paper)
                }

                return target, nil
            }

        // Message Array
        // OI
        case "QueryMessageWithPaperKey" :
            if len(target) == 0 {
                return nil, errors.New("ApplyAcceptanceModel" + " : " + "The comment does not exist.")
            }

            identifier, err = GetIdentifier(stub)
            if err != nil {
                return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
            }

            if userType == EOP {
                return target, nil
            } else if userType == ROP {
                err = json.Unmarshal(target, &messageArray)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                if len(messageArray) == 0 {
                    return target, nil
                }

                paper, err = GetPaper(stub, messageArray[0].PaperKey)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                if OICheck(paper, acceptanceModel) {
                    return target, nil
                } else {
                    buff.WriteString("[")

                    for _, message := range messageArray {
                        if message.ReviewerID != identifier {
                            message.ReviewerID = ""
                        }

                        messageByte, err = json.Marshal(message)
                        if err != nil {
                            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                        }

                        if writeOne {
                            buff.WriteString(",")
                        }

                        writeOne = true
                        buff.Write(messageByte)
                    }

                    buff.WriteString("]")

                    return buff.Bytes(), nil
                }
            } else {
                err = json.Unmarshal(target, &messageArray)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                if len(messageArray) == 0 {
                    return target, nil
                }

                paper, err = GetPaper(stub, messageArray[0].PaperKey)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                if OICheck(paper, acceptanceModel) {
                    return target, nil
                } else {
                    buff.WriteString("[")

                    for _, message := range messageArray {
                        message.ReviewerID = ""
                        messageByte, err = json.Marshal(message)
                        if err != nil {
                            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                        }

                        if writeOne {
                            buff.WriteString(",")
                        }

                        writeOne = true
                        buff.Write(messageByte)
                    }

                    buff.WriteString("]")

                    return buff.Bytes(), nil
                }
            }

        case "QueryMessageWithContractKey" :
            if len(target) == 0 {
                return nil, errors.New("ApplyAcceptanceModel" + " : " + "The comment does not exist.")
            }

            identifier, err = GetIdentifier(stub)
            if err != nil {
                return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
            }

            if userType == EOP {
                return target, nil
            } else if userType == ROP {
                err = json.Unmarshal(target, &messageArray)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                if len(messageArray) == 0 {
                    return target, nil
                }

                paper, err = GetPaper(stub, messageArray[0].PaperKey)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                if OICheck(paper, acceptanceModel) {
                    return target, nil
                } else {
                    buff.WriteString("[")

                    for _, message := range messageArray {
                        if message.ReviewerID != identifier {
                            message.ReviewerID = ""
                        }

                        messageByte, err = json.Marshal(message)
                        if err != nil {
                            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                        }

                        if writeOne {
                            buff.WriteString(",")
                        }

                        writeOne = true
                        buff.Write(messageByte)
                    }

                    buff.WriteString("]")

                    return buff.Bytes(), nil
                }
            } else {
                err = json.Unmarshal(target, &messageArray)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                if len(messageArray) == 0 {
                    return target, nil
                }

                paper, err = GetPaper(stub, messageArray[0].PaperKey)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                if OICheck(paper, acceptanceModel) {
                    return target, nil
                } else {
                    buff.WriteString("[")

                    for _, message := range messageArray {
                        message.ReviewerID = ""
                        messageByte, err = json.Marshal(message)
                        if err != nil {
                            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                        }

                        if writeOne {
                            buff.WriteString(",")
                        }

                        writeOne = true
                        buff.Write(messageByte)
                    }

                    buff.WriteString("]")

                    return buff.Bytes(), nil
                }
            }
        // ON
        case "QueryMessageWithReviewerID" :
            if len(target) == 0 {
                return nil, errors.New("ApplyAcceptanceModel" + " : " + "The comment does not exist.")
            }

            if userType == SELF {
                return target, nil
            }

            identifier, err = GetIdentifier(stub)
            if err != nil {
                return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
            }

            err = json.Unmarshal(target, &messageArray)
            if err != nil {
                return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
            }

            buff.WriteString("[")

            for _, message := range messageArray {
                paper, err = GetPaper(stub, message.PaperKey)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                userType, err = GetUserTypeWithPaper(stub, paper)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                if userType == EOP {
                    if writeOne {
                        buff.WriteString(",")
                    }

                    writeOne = true
                    buff.Write(messageByte)
                } else if userType == ROP {
                    acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    if !OICheck(paper, acceptanceModel) {
                        message.ReviewerID = ""
                        messageByte, err = json.Marshal(message)
                    }

                    if writeOne {
                        buff.WriteString(",")
                    }

                    writeOne = true
                    buff.Write(messageByte)
                } else {
                    acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    if ONCheck(paper, acceptanceModel) {
                        if !OICheck(paper, acceptanceModel) {
                            message.ReviewerID = ""
                            messageByte, err = json.Marshal(message)
                        }

                        if writeOne {
                            buff.WriteString(",")
                        }

                        writeOne = true
                        buff.Write(messageByte)
                    }
                }
            }

            buff.WriteString("]")

            return buff.Bytes(), nil
        }

    return target, nil
}

func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
    var args []string
    var userType UserType
    var result []byte
    var maskedResult []byte
    var err error

    //get function name and argument list
    function, args = stub.GetFunctionAndParameters()

    //access control
    userType, err = AccessControl(stub, function, args)
    if err != nil {
        return shim.Error("message : " + err.Error())
    }

    switch function {
        // No Return
        case "AddMessage" :
            err = s.AddMessage(stub, args, userType)
        case "InitMessage" :
            err = s.InitMessage(stub, args)
        case "DeleteMessageWithPaperKey" :
            err = s.DeleteMessageWithPaperKey(stub, args)

        // Message
        case "QueryMessage" :
            result, err = s.QueryMessage(stub, args)

        // Message Array
        case "QueryMessageWithPaperKey" :
            if userType == AERROR {
                return shim.Success([]byte("[]"))
            }
            result, err = s.QueryMessageWithPaperKey(stub, args)
        case "QueryMessageWithContractKey" :
            if userType == AERROR {
                return shim.Success([]byte("[]"))
            }
            result, err = s.QueryMessageWithContractKey(stub, args)
        case "QueryMessageWithReviewerID" :
            if userType == AERROR {
                return shim.Success([]byte("[]"))
            }
            result, err = s.QueryMessageWithReviewerID(stub, args)
    }

    if err != nil {
        return shim.Error("message : " + err.Error())
    }

    if result != nil {
        maskedResult, err = ApplyAcceptanceModel(stub, function, userType, result)
        if err != nil {
            return shim.Error("message : " + err.Error())
        }
    }

    return shim.Success(maskedResult)
}


//args = [ contractKey ]
func (s *SmartContract) InitMessage(stub shim.ChaincodeStubInterface, args []string) error {
    var key string
    var err error

    if len(args) != 1 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    key, err = CreateMessageIndexKey(stub, args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(key, []byte("0"))
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

//args = [ parent, message ]
func (s *SmartContract) AddMessage(stub shim.ChaincodeStubInterface, args []string, userType UserType) error {
    var identifier string
    var indexKey string
    var paperKey string
    var reviewerKey string
    var reviewerKeyByte []byte
    var roundString string
    var messageByte []byte
    var message Message
    var indexByte []byte
    var indexString string
    var index int
    var key string
    var RMKey string
    var err error

    if len(args) != 2 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 2")
    }

    paperKey, err = CreatePaperKeyWithContractKey(stub, args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    indexKey, err = CreateMessageIndexKey(stub, args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    indexByte, err = stub.GetState(indexKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    index, err = strconv.Atoi(string(indexByte))
    if err != nil {
        return errors.New(function + " : " + indexKey + ", " + err.Error())
    }
    index = index + 1
    indexString = strconv.Itoa(index)

    key, err = CreateMessageKey(stub, args[0], indexString)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    identifier, err = GetIdentifier(stub)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    roundString, err = GetRoundFromContractKey(stub, args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    if userType == ROP {
        reviewerKeyByte, err = InvokeChaincode(stub, "reviewer", "GetPRKeyWithPaperKey", []string{ paperKey, identifier })
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        reviewerKey = string(reviewerKeyByte)
    } else if userType == AUTHOR {
        reviewerKey = "AUTHOR"
    } else if userType == EOP {
        reviewerKey = "EOP"
    } else {
        reviewerKey = "REVIEWER"
    }

    updateTime, err := stub.GetTxTimestamp()
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    message = Message { Key : key, ReviewerKey : reviewerKey, ReviewerID : identifier, PaperKey : paperKey, Round : roundString, ContractKey : args[0], Message : args[1], Date : time.Unix(updateTime.GetSeconds(), 0).String() }

    messageByte, err = json.Marshal(message)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(key, messageByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(args[0], []byte(indexString))
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    RMKey, err = ConvertMessageKey(stub, key, identifier)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(RMKey, []byte(key))
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(indexKey, []byte(indexString))
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    indexKey, err = CreateMessageIndexKey(stub, key)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(indexKey, []byte("0"))
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

//args = [ paperKey ]
func (s *SmartContract) DeleteMessageWithPaperKey(stub shim.ChaincodeStubInterface, args []string) error {
    var messageIter shim.StateQueryIteratorInterface
    var splitPaperKey []string
    var err error

    if len(args) != 1 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    _, splitPaperKey, err = stub.SplitCompositeKey(args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    messageIter, err = stub.GetStateByPartialCompositeKey("", splitPaperKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if !messageIter.HasNext() {
        return nil
    }

    for messageIter.HasNext() {
        messageKV, err := messageIter.Next()
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        err = stub.DelState(messageKV.Key)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }
    }

    return nil
}

//args = [ messageKey ]
func (s *SmartContract) QueryMessage(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var messageByte []byte
    var err error

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    messageByte, err = stub.GetState(args[0])
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if len(messageByte) == 0 {
        return nil, errors.New(function + " : " + args[0] + " message does not exist.")
    }

    return messageByte, nil
}

//args = [ paperKey ]
func (s *SmartContract) QueryMessageWithPaperKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var splitPaperKey []string
    var messageIter shim.StateQueryIteratorInterface
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

    messageIter, err = stub.GetStateByPartialCompositeKey("MR", splitPaperKey)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !messageIter.HasNext() {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for messageIter.HasNext() {
        messageKV, err := messageIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if flag {
            buff.WriteString(",")
        } else {
          flag = true
        }

        buff.Write(messageKV.Value)
    }

    buff.WriteString("]")
    messageIter.Close()

    return buff.Bytes(), nil
}

//args = [ contractKey ]
func (s *SmartContract) QueryMessageWithContractKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var splitContractKey []string
    var messageIter shim.StateQueryIteratorInterface
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

    messageIter, err = stub.GetStateByPartialCompositeKey("MR", splitContractKey)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !messageIter.HasNext() {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for messageIter.HasNext() {
        messageKV, err := messageIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if flag {
            buff.WriteString(",")
        } else {
          flag = true
        }

        buff.Write(messageKV.Value)
    }

    buff.WriteString("]")
    messageIter.Close()

    return buff.Bytes(), nil
}

//args = [ ReviewerID ]
func (s *SmartContract) QueryMessageWithReviewerID(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var MRKeyIter shim.StateQueryIteratorInterface
    var messageByte []byte
    var buff bytes.Buffer
    var flag bool
    var err error

    flag = false

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    MRKeyIter, err = stub.GetStateByPartialCompositeKey("RM", []string{ args[0] })
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !MRKeyIter.HasNext() {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for MRKeyIter.HasNext() {
        RMKey, err := MRKeyIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        messageByte, err = stub.GetState(string(RMKey.Value))
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }
        if len(messageByte) == 0 {
            return nil, errors.New(string(RMKey.Value))
        }

        if flag {
            buff.WriteString(",")
        } else {
          flag = true
        }

        buff.Write(messageByte)
    }

    buff.WriteString("]")
    MRKeyIter.Close()

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
