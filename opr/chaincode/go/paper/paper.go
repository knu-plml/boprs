package main
import (
    "fmt"
    "bytes"
    "encoding/json"
    "time"
    "strconv"
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
    var paperByte []byte
    var paper Paper
    var userType UserType
    var acceptanceModel *AcceptanceModel
    var err error

    userType = REVIEWER

    switch function {
        case "GetIdentifier", "GetEmail" :
            return SELF, nil

        //Editor of Organization args[0] = organization
        case "AddPaper" :
            organization, err := GetMSPID(stub)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            if organization == args[0] {
                userType = EOO
            } else {
                return userType, errors.New("AccessControl"+ " : " + "You are not editor of " + organization + ".")
            }


            return userType, nil
        case "AddPaperWithEditorID" :
            err = CertifyBroker(stub)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            } else {
                userType = EOO
            }

            return userType, nil
        //Editor of Paper or broker
        //args[0] = paperKey
        case "UpdateStatus" :
            err = CertifyBroker(stub)
            if err == nil {
                userType = EOP
                return userType, nil
            }

            paperByte, err = stub.GetState(args[0])
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }
            if len(paperByte) == 0 {
                return userType, errors.New("AccessControl"+ " : " + "There is no paper with the Key " + args[0] + ".")
            }

            err = json.Unmarshal(paperByte, &paper)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            err = CertifyEOPWithPaper(stub, &paper)
            if err == nil {
                userType = EOP
            } else {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            if paper.Status == "accepted" && paper.Status == "rejected" {
                return userType, errors.New("AccessControl" + " : " + "Can't update Round because the status is " + paper.Status + ".")
            }

            return userType, nil
        case "UpdateRound" :
            paperByte, err = stub.GetState(args[0])
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }
            if len(paperByte) == 0 {
                return userType, errors.New("AccessControl"+ " : " + "There is no paper with the Key " + args[0] + ".")
            }

            err = json.Unmarshal(paperByte, &paper)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            err = CertifyEOPWithPaper(stub, &paper)
            if err == nil {
                userType = EOP
            } else {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            return userType, nil
        //Anyone args[0] = paperKey
        case "QueryPaper" :
            paperByte, err = stub.GetState(args[0])
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }
            if len(paperByte) == 0 {
                return userType, errors.New("AccessControl"+ " : " + "There is no paper with the Key " + args[0] + ".")
            }

            err = json.Unmarshal(paperByte, &paper)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            userType, err = GetUserTypeWithPaper(stub, &paper)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            return userType, nil
        //Anyone args[0] = paperKey
        case "QueryPaperFileWithPaperKey", "QueryCurrentPaperFile" :
            paperByte, err = stub.GetState(args[0])
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }
            if len(paperByte) == 0 {
                return userType, errors.New("AccessControl"+ " : " + "There is no paper with the Key " + args[0] + ".")
            }

            err = json.Unmarshal(paperByte, &paper)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            userType, err = GetUserTypeWithPaper(stub, &paper)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            if userType == EDITOR || userType == REVIEWER {
                acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                if err != nil {
                    return userType, errors.New("AccessControl"+ " : " + err.Error())
                }

                if acceptanceModel.OM == "0" && paper.Status != "accepted" && paper.Status != "rejected" {
                    return userType, errors.New("AccessControl"+ " : " + "AcceptanceModel Error, Open pre-review Manuscripts")
                }
            }

            return userType, nil
        // Anyone args[0] = authorID
        case "QueryPaperWithAuthorID" :
            err = CertifyIdentifier(stub, args[0])
            if err == nil {
                userType = AUTHOR
            }

            err = CertifyEditor(stub)
            if err == nil {
                userType = EDITOR
            }

            return userType, nil
        // Editor himself args[0] = editorID
        case "QueryPaperWithEditorID" :
            err = CertifyIdentifier(stub, args[0])
            if err == nil {
                userType = EOP
            } else {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            return userType, nil
        // Anyone args[0] = organization 
        case "QueryPaperWithOrganization" :
            organization, err := GetMSPID(stub)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            if organization == args[0] {
                userType = EOO
            }

            return userType, nil
        // Anyone args[0] = paperFileKey
        case "QueryPaperFile" :
            paperKey, err := CreatePaperKeyWithPaperFileKey(stub, args[0])
            paperByte, err = stub.GetState(paperKey)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }
            if len(paperByte) == 0 {
                return userType, errors.New("AccessControl"+ " : " + "There is no paper with the Key " + args[0] + ".")
            }

            err = json.Unmarshal(paperByte, &paper)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            userType, err = GetUserTypeWithPaper(stub, &paper)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            if userType == EDITOR || userType == REVIEWER {
                acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                if err != nil {
                    return userType, errors.New("AccessControl"+ " : " + err.Error())
                }

                if acceptanceModel.OM == "0" && paper.Status != "accepted" && paper.Status != "rejected" {
                    return userType, errors.New("AccessControl"+ " : " + "AcceptanceModel Error, Open pre-review Manuscripts")
                }
            }

            return userType, nil
        // Anyone
        case "QueryAllPaper" :
            err = CertifyEditor(stub)
            if err == nil {
                userType = EDITOR
            }

            return userType, nil
        //Author or Editor of paper 
        //args[0] = paperKey
        case "QueryPaperHistory", "UpdatePaper", "DeletePaper", "DeleteRelatedData" :
            paperByte, err = stub.GetState(args[0])
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }
            if len(paperByte) == 0 {
                return userType, errors.New("AccessControl"+ " : " + "There is no paper with the Key " + args[0] + ".")
            }

            err = json.Unmarshal(paperByte, &paper)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            err = CertifyEOPWithPaper(stub, &paper)
            if err == nil {
                userType = EOP
            } else {
                err = CertifyAuthorWithPaper(stub, &paper)
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

func ApplyAcceptanceModel(stub shim.ChaincodeStubInterface, function string, userType UserType, target []byte) ([]byte, error) {
    var paper Paper
    var paperByte []byte
    var paperArray []Paper
    var paperFile PaperFile
    var paperFileArray []PaperFile
    var acceptanceModel *AcceptanceModel
    var buff bytes.Buffer
    var result []byte
    var writeOne bool
    var err error


    result = target
    writeOne = false

    //return Paper
    if function == "QueryPaper" {
        err = json.Unmarshal(target, &paper)
        if err != nil {
            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
        }

        acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
        if err != nil {
            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
        }

        if userType != AUTHOR && userType != EOP && !OICheck(&paper, acceptanceModel) {
            paper.AuthorID = ""
            paper.AuthorEmail = ""
            result, err = json.Marshal(paper)
            if err != nil {
                return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
            }
        }

        return result, nil
    } else if function == "QueryPaperFile" || function == "QueryCurrentPaperFile" || function == "QueryPaperFileWithPaperKey" {
        if userType == EOP || userType == EOO || userType == AUTHOR || userType == ROP {
            return target, nil
        }

        if function == "QueryPaperFileWithPaperKey" {
            err = json.Unmarshal(target, &paperFileArray)
            if err != nil {
                return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
            }
            if len(paperFileArray) == 0 {
                return target, nil
            }

            paperFile = paperFileArray[0]
        } else {
            err = json.Unmarshal(target, &paperFile)
            if err != nil {
                return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
            }
        }

        paperKey, err := CreatePaperKeyWithPaperFileKey(stub, paperFile.Key)
        if err != nil {
            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
        }

        paperByte, err := stub.GetState(paperKey)
        if err != nil {
            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
        }

        err = json.Unmarshal(paperByte, &paper)
        if err != nil {
            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
        }

        acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
        if err != nil {
            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
        }

        if acceptanceModel.OM == "1" || (acceptanceModel.OM == "0" && (paper.Status == "accepted" || paper.Status == "rejected")) {
            return target, nil
        } else {
            return nil, nil
        }
    } else {
        //return Paper Array
        err = json.Unmarshal(result, &paperArray)
        if err != nil {
            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
        }
        if len(paperArray) == 0 {
            return target, nil
        }

        switch function {
            case "QueryAllPaper" :
                buff.WriteString("[")

                for _, paper := range paperArray {
                    acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    userType, err = GetUserTypeWithPaper(stub, &paper)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    if userType != EIC && userType != EOP && userType != AUTHOR && !OICheck(&paper, acceptanceModel) {
                        paper.AuthorID = ""
                        paper.AuthorEmail = ""
                    }

                    paperByte, err = json.Marshal(paper)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    if writeOne {
                        buff.WriteString(",")
                    }

                    writeOne = true
                    buff.Write(paperByte)
                }

                buff.WriteString("]")

                return buff.Bytes(), nil

            case "QueryPaperWithAuthorID" :
                buff.WriteString("[")
                if userType == AUTHOR {
                    return target, nil
                } else if userType == EDITOR {
                    for _, paper := range paperArray {
                        organization, err := GetMSPID(stub)
                        if err != nil {
                            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                        }

                        if organization == paper.Organization {
                            paperByte, err = json.Marshal(paper)
                            if err != nil {
                                return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                            }

                            if writeOne {
                                buff.WriteString(",")
                            }

                            writeOne = true
                            buff.Write(paperByte)
                        }
                    }

                    buff.WriteString("]")

                    return buff.Bytes(), nil
                } else {
                    for _, paper := range paperArray {
                        acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                        if err != nil {
                            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                        }
                        if OICheck(&paper, acceptanceModel) {
                            paperByte, err = json.Marshal(paper)
                            if err != nil {
                                return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                            }

                            if writeOne {
                                buff.WriteString(",")
                            }

                            writeOne = true
                            buff.Write(paperByte)
                        }
                    }

                    buff.WriteString("]")

                    return buff.Bytes(), nil
                }

            case "QueryPaperWithOrganization" :
                if userType == EOO || userType == EOP {
                    return target, nil
                } else {
                    if len(paperArray) == 0 {
                        return target, nil
                    }

                    buff.WriteString("[")
                    for _, paper := range paperArray {
                        acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                        if err != nil {
                            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                        }

                        userType, err = GetUserTypeWithPaper(stub, &paper)
                        if err != nil {
                            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                        }

                        if userType != EOP && userType != AUTHOR && !OICheck(&paper, acceptanceModel) {
                            paper.AuthorID = ""
                            paper.AuthorEmail = ""
                        }

                        paperByte, err = json.Marshal(paper)
                        if err != nil {
                            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                        }

                        if writeOne {
                            buff.WriteString(",")
                        }

                        writeOne = true
                        buff.Write(paperByte)
                    }

                    buff.WriteString("]")

                    return buff.Bytes(), nil
                }

            case "QueryPaperWithEditorID", "QueryPaperHistory" :
                return target, nil

            default :
                return nil, errors.New("ApplyAcceptanceModel" + " : " + function + " is Invalid Smart Contract function name.")
        }
    }
}

func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
    var args []string
    var userType UserType
    var result []byte = nil
    var maskedResult []byte = nil
    var err error

    //get function name and argument list
    function, args = stub.GetFunctionAndParameters()

    if function == "CertifyEditor" {
        err = CertifyEditor(stub)
        if err != nil {
            return shim.Success([]byte("false"))
        } else {
            return shim.Success([]byte("true"))
        }
    } else if function == "CertifyEIC" {
        err = CertifyEIC(stub, args[0])
        if err != nil {
            return shim.Success([]byte("false"))
        } else {
            return shim.Success([]byte("true"))
        }
    } else if function == "GetIdentifier" {
        identifier, err := GetIdentifier(stub)
        if err != nil {
            return shim.Error("paper : " + err.Error())
        }

        return shim.Success([]byte(identifier))
    } else if function == "GetEmail" {
        identifier, err := GetEmail(stub)
        if err != nil {
            return shim.Error("paper : " + err.Error())
        }

        return shim.Success([]byte(identifier))
    }

    //access control
    userType, err = AccessControl(stub, function, args)
    if err != nil {
        return shim.Error("paper : " + err.Error())
    }

    switch function {
        //return Paper Array
        case "QueryAllPaper" :
            result, err = s.QueryAllPaper(stub, args)
        case "QueryPaperWithAuthorID" :
            result, err = s.QueryPaperWithAuthorID(stub, args)
        case "QueryPaperWithEditorID" :
            result, err = s.QueryPaperWithEditorID(stub, args)
        case "QueryPaperWithOrganization" :
            result, err = s.QueryPaperWithOrganization(stub, args)
        case "QueryPaperHistory" :
            result, err = s.QueryPaperHistory(stub, args)

        //return PaperFile
        case "QueryPaperFileWithPaperKey" :
            result, err = s.QueryPaperFileWithPaperKey(stub, args)
        case "QueryCurrentPaperFile" :
            result, err = s.QueryCurrentPaperFile(stub, args)
        case "QueryPaperFile" :
            result, err = s.QueryPaperFile(stub, args)

        //return Paper
        case "QueryPaper" :
            result, err = s.QueryPaper(stub, args)

        //No Return
        case "AddPaper" :
            err = s.AddPaper(stub, args)
        case "AddPaperWithEditorID" :
            err = s.AddPaperWithEditorID(stub, args)
        case "UpdatePaper" :
            err = s.UpdatePaper(stub, args)
        case "UpdateStatus" :
            err = s.UpdateStatus(stub, args)
        case "UpdateRound" :
            err = s.UpdateRound(stub, args)
        case "DeletePaper" :
            err = s.DeletePaper(stub, args)
        case "DeleteRelatedData" :
            err = s.DeleteRelatedData(stub, args)
    }

    if err != nil {
        return shim.Error("paper : " + err.Error())
    }

//    maskedResult = result
    if result != nil {
        maskedResult, err = ApplyAcceptanceModel(stub, function, userType, result)
        if err != nil {
            return shim.Error("paper : " + err.Error())
        }
    }

    return shim.Success(maskedResult)
}

// args = [ paperKey ]
func (s *SmartContract) QueryPaper(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var paperByte []byte
    var err error

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    paperByte, err = stub.GetState(args[0])
    if err != nil {
        return nil, errors.New(function + " : " + "Error occured in quering aper")
    }
    if len(paperByte) == 0 {
        return nil, errors.New(function + " : " + args[0] + " paper does not exist.")
    }
    return paperByte, nil
}

// args = [ ]
func (s *SmartContract) QueryAllPaper(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var paperIter shim.StateQueryIteratorInterface
    var buff bytes.Buffer
    var flag bool
    var err error

    flag = false

    if len(args) != 0 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 0")
    }

    paperIter, err = stub.GetStateByPartialCompositeKey("OI", []string { })
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !paperIter.HasNext() {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for paperIter.HasNext() {
        paperKV, err := paperIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if flag {
            buff.WriteString(",")
        } else {
          flag = true
        }

        buff.Write(paperKV.Value)
    }

    buff.WriteString("]")
    paperIter.Close()

    return buff.Bytes(), nil
}

// args = [ authorID ]
func (s *SmartContract) QueryPaperWithAuthorID(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var OIKeyIter shim.StateQueryIteratorInterface
    var paperByte []byte
    var buff bytes.Buffer
    var flag bool
    var err error

    flag = false

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    OIKeyIter, err = stub.GetStateByPartialCompositeKey("AOI", []string { args[0] })
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !OIKeyIter.HasNext() {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for OIKeyIter.HasNext() {
        OIKeyKV, err := OIKeyIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        paperByte, err = stub.GetState(string(OIKeyKV.Value))
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if flag {
            buff.WriteString(",")
        } else {
            flag = true
        }

        buff.Write(paperByte)
    }

    buff.WriteString("]")
    OIKeyIter.Close()

    return buff.Bytes(), nil
}

// args = [ editorID ]
func (s *SmartContract) QueryPaperWithEditorID(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var paperIter shim.StateQueryIteratorInterface
    var paper Paper
    var buff bytes.Buffer
    var flag bool
    var err error

    flag = false

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    paperIter, err = stub.GetStateByPartialCompositeKey("OI", []string{ })
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !paperIter.HasNext() {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for paperIter.HasNext() {
        paperKV, err := paperIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        err = json.Unmarshal(paperKV.Value, &paper)
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }
        if paper.EditorID == args[0] {
            if flag {
                buff.WriteString(",")
            } else {
              flag = true
            }
            buff.Write(paperKV.Value)
        }
    }

    buff.WriteString("]")
    paperIter.Close()

    return buff.Bytes(), nil
}

//args = [ organization ]
func (s *SmartContract) QueryPaperWithOrganization(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var paperIter shim.StateQueryIteratorInterface
    var buff bytes.Buffer
    var flag bool
    var err error

    flag = false

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    paperIter, err = stub.GetStateByPartialCompositeKey("OI", []string { args[0] })
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !paperIter.HasNext() {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for paperIter.HasNext() {
        paperKV, err := paperIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if flag {
            buff.WriteString(",")
        } else {
            flag = true
        }

        buff.Write(paperKV.Value)
    }

    buff.WriteString("]")
    paperIter.Close()

    return buff.Bytes(), nil
}

// args = [ paperKey ]
func (s *SmartContract) QueryPaperHistory(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var paperIter shim.HistoryQueryIteratorInterface
    var paper Paper
    var paperHistory *PaperHistory
    var paperHistoryByte []byte
    var buff bytes.Buffer
    var flag bool
    var err error

    flag = false

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    paperIter, err = stub.GetHistoryForKey(args[0])
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if(!paperIter.HasNext()) {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for paperIter.HasNext() {
        history, err := paperIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        err = json.Unmarshal(history.Value, &paper)
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        paperHistory = &PaperHistory { Paper : paper, Timestamp : time.Unix(history.GetTimestamp().Seconds, 0).String() }
        paperHistoryByte, err = json.Marshal(paperHistory)
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if flag {
            buff.WriteString(",")
        } else {
            flag = true
        }

        buff.Write(paperHistoryByte)
    }

    buff.WriteString("]")
    paperIter.Close()

    return buff.Bytes(), nil
}

// args = [ paperFileKey ]
func (s *SmartContract) QueryPaperFile(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var paperFileByte []byte
    var err error

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    paperFileByte, err = stub.GetState(args[0])
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if len(paperFileByte) == 0 {
        return nil, errors.New(function + " : " + args[0] + " paper does not exist.")
    }

    return paperFileByte, nil
}

// args = [ paperKey ]
func (s *SmartContract) QueryPaperFileWithPaperKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var splitedPaperKey []string
    var paperFileIter shim.StateQueryIteratorInterface
    var buff bytes.Buffer
    var flag bool
    var err error

    flag = false

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    _, splitedPaperKey, err = stub.SplitCompositeKey(args[0])
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }

    paperFileIter, err = stub.GetStateByPartialCompositeKey("PF", splitedPaperKey)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !paperFileIter.HasNext() {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for paperFileIter.HasNext() {
        paperFileKV, err := paperFileIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if flag {
            buff.WriteString(",")
        } else {
          flag = true
        }

        buff.Write(paperFileKV.Value)
    }

    buff.WriteString("]")
    paperFileIter.Close()

    return buff.Bytes(), nil
}

// args = [ paperKey ]
func (s *SmartContract) QueryCurrentPaperFile(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var paperByte []byte
    var paper Paper
    var paperFileKey string
    var paperFileByte []byte
    var err error

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    paperByte, err = stub.GetState(args[0])
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if len(paperByte) == 0 {
        return nil, errors.New(function + " : " + args[0] + " paper does not exist.")
    }

    err = json.Unmarshal(paperByte, &paper)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }

    paperFileKey, err = CreatePaperFileKeyWithPaperKey(stub, args[0], paper.Round + 1)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }

    paperFileByte, err = stub.GetState(paperFileKey)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if len(paperFileByte) == 0 {
        paperFileKey, err = CreatePaperFileKeyWithPaperKey(stub, args[0], paper.Round)
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        paperFileByte, err = stub.GetState(paperFileKey)
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }
        if len(paperFileByte) == 0 {
            return nil, errors.New(function + " : " + paperFileKey + " paper file does not exist.")
        }
    }

    return paperFileByte, nil
}

// args = [ organization, authorID, paperID, paperTitle, paper, email, paperAbstract ]
func (s *SmartContract) AddPaper(stub shim.ChaincodeStubInterface, args []string) error {

    var editorID string
    var OIKey string
    var AOIKey string
    var paper Paper
    var paperByte []byte
    var paperFileKey string
    var paperFile PaperFile
    var paperFileByte []byte
    var err error

    if len(args) != 7 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 7")
    }

    if len(args[0]) == 0 || len(args[1]) == 0 || len(args[2]) == 0 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 7")
    }


    editorID, err = GetIdentifier(stub)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    OIKey, err = CreatePaperKey(stub, args[0], args[2])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    AOIKey, err = ConvertPaperKey(stub, OIKey, args[1])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    paperByte, err = stub.GetState(OIKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if len(paperByte) != 0 {
        return errors.New(function + " : " + OIKey + " paper already exists.")
    }

    paper = Paper { Key : OIKey, Organization : args[0], AuthorID : args[1], PaperID : args[2], Title : args[3], AuthorEmail : args[5], Abstract : args[6], EditorID : editorID, Round : 0, Status : "recruit_reviewer", ContractKey : "" }

    paperByte, err = json.Marshal(paper)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(OIKey, paperByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(AOIKey, []byte(OIKey))
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    paperFileKey, err = CreatePaperFileKeyWithPaperKey(stub, OIKey, 1)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    updateTime, err := stub.GetTxTimestamp()
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    paperFile = PaperFile { Key : paperFileKey, File : args[4], RevisionNote : "", Date : time.Unix(updateTime.GetSeconds(), 0).String() }

    paperFileByte, err = json.Marshal(paperFile)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(paperFileKey, paperFileByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

// args = [ organization, authorID, paperID, paperTitle, paper, authorEmail, paperAbstract, editorID ]
func (s *SmartContract) AddPaperWithEditorID(stub shim.ChaincodeStubInterface, args []string) error {

    var editorID string
    var OIKey string
    var AOIKey string
    var paper Paper
    var paperByte []byte
    var paperFileKey string
    var paperFile PaperFile
    var paperFileByte []byte
    var err error

    if len(args) != 8 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 8")
    }

    if len(args[0]) == 0 || len(args[1]) == 0 || len(args[2]) == 0 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 8")
    }

    editorID = args[7]

    OIKey, err = CreatePaperKey(stub, args[0], args[2])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    AOIKey, err = ConvertPaperKey(stub, OIKey, args[1])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    paperByte, err = stub.GetState(OIKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if len(paperByte) != 0 {
        return errors.New(function + " : " + OIKey + " paper already exists.")
    }

    paper = Paper { Key : OIKey, Organization : args[0], AuthorID : args[1], PaperID : args[2], Title : args[3], AuthorEmail : args[5], Abstract : args[6], EditorID : editorID, Round : 0, Status : "recruit_reviewer", ContractKey : "" }

    paperByte, err = json.Marshal(paper)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(OIKey, paperByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(AOIKey, []byte(OIKey))
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    paperFileKey, err = CreatePaperFileKeyWithPaperKey(stub, OIKey, 1)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    updateTime, err := stub.GetTxTimestamp()
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    paperFile = PaperFile { Key : paperFileKey, File : args[4], RevisionNote : "", Date : time.Unix(updateTime.GetSeconds(), 0).String() }

    paperFileByte, err = json.Marshal(paperFile)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(paperFileKey, paperFileByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

// args = [ paperKey, paperFile, revisionNote ]
func (s *SmartContract) UpdatePaper(stub shim.ChaincodeStubInterface, args []string) error {
    var paper Paper
    var paperByte []byte
    var paperFileKey string
    var paperFile PaperFile
    var paperFileByte []byte
    var err error

    if len(args) != 3 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 3")
    }

    paperByte, err = stub.GetState(args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if len(paperByte) == 0 {
        return errors.New(function + " : " + args[0] + " paper does not exist.")
    }

    err = json.Unmarshal(paperByte, &paper)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    if paper.Status != "recruit_reviewer" && paper.Status != "revise" {
        return errors.New(function + " : " + args[0] + " paper can not be updated. It must be in the 'revise' or 'recruit_reviewer' state.")
    }

    paperFileKey, err = CreatePaperFileKeyWithPaperKey(stub, args[0], paper.Round + 1)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    updateTime, err := stub.GetTxTimestamp()
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    paperFile = PaperFile { Key : paperFileKey, File : args[1], RevisionNote : args[2], Date : time.Unix(updateTime.GetSeconds(), 0).String() }
    paperFileByte, err = json.Marshal(paperFile)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(paperFileKey, paperFileByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

// args = [ paperKey, status ]
func (s *SmartContract) UpdateStatus(stub shim.ChaincodeStubInterface, args []string) error {
    var paper Paper
    var paperByte []byte
    var err error

    if len(args) != 2 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 2")
    }

    paperByte, err = stub.GetState(args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if len(paperByte) == 0 {
        return errors.New(function + " : " + args[0] + " paper does not exist.")
    }

    err = json.Unmarshal(paperByte, &paper)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    paper.Status = args[1]

    paperByte, err = json.Marshal(paper)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(args[0], paperByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

// args = [ paperKey, round ]
func (s *SmartContract) UpdateRound(stub shim.ChaincodeStubInterface, args []string) error {
    var paper Paper
    var paperByte []byte
    var contractKey string
    var err error

    if len(args) != 2 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 2")
    }

    paperByte, err = stub.GetState(args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if len(paperByte) == 0 {
        return errors.New(function + " : " + args[0] + " paper does not exist.")
    }

    err = json.Unmarshal(paperByte, &paper)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    contractKey, err = CreateContractKeyWithPaperKey(stub, args[0], args[1])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    paper.ContractKey = contractKey
    paper.Round, err = strconv.Atoi(args[1])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    paper.Status = "reviewer_invited"

    paperByte, err = json.Marshal(paper)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(args[0], paperByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

// args = [ paperKey ]
func (s *SmartContract) DeletePaper(stub shim.ChaincodeStubInterface, args []string) error {
    var AOIKey string
    var paper Paper
    var paperByte []byte
    var splitedPaperKey []string
    var paperFileIter shim.StateQueryIteratorInterface
    var err error

    if len(args) != 1 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    paperByte, err = stub.GetState(args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if len(paperByte) == 0 {
        return errors.New(function + " : " + args[0] + " paper does not exist.")
    }

    err = json.Unmarshal(paperByte, &paper)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    AOIKey, err = ConvertPaperKey(stub, args[0], paper.AuthorID)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.DelState(args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.DelState(AOIKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    _, splitedPaperKey, err = stub.SplitCompositeKey(args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    paperFileIter, err = stub.GetStateByPartialCompositeKey("PF", splitedPaperKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if !paperFileIter.HasNext() {
        return nil
    }

    for paperFileIter.HasNext() {
        paperFileKV, err := paperFileIter.Next()
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        err = stub.DelState(paperFileKV.Key)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }
    }
    paperFileIter.Close()

    return nil
}

// args = [ paperKey ]
func (s *SmartContract) DeleteRelatedData(stub shim.ChaincodeStubInterface, args []string) error {
    var err error

    _, err = InvokeChaincode(stub, "message", "DeleteMessageWithPaperKey", args)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    _, err = InvokeChaincode(stub, "comment", "DeleteCommentWithPaperKey", args)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    _, err = InvokeChaincode(stub, "report", "DeleteReportWithPaperKey", args)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    _, err = InvokeChaincode(stub, "contract", "DeleteContractWithPaperKey", args)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    _, err = InvokeChaincode(stub, "reviewer", "DeleteReviewerWithPaperKey", args)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    err = s.DeletePaper(stub, args)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}


// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

    // Create a new Smart Contract
    err := shim.Start(new(SmartContract))
    if err != nil {
        fmt.Printf("Error creating new Smart Contract: %s", err)
    }
}
