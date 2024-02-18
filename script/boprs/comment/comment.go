package main
import (
    "encoding/json"
    "fmt"
    "bytes"
    "errors"
    "strconv"

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
    var userID string
    var userType UserType
    var paperKey string
    var paper *Paper
    var commentKey string
    var acceptanceModel *AcceptanceModel
    var err error

    userType = REVIEWER
    userID, err = GetIdentifier(stub)
    if err != nil {
        return userType, errors.New("AccessControl"+ " : " + err.Error())
    }

    switch function {
        //reviewer of paper
        // args[0] = contractKey
        case "AddComment" :
            paperKey, err = CreatePaperKeyWithContractKey(stub, args[0])
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            paper, err = GetPaper(stub, paperKey)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            err = CertifyROPWithPaperKey(stub, paperKey)
            if err == nil {
                reviewerKeyByte, err := InvokeChaincode(stub, "reviewer", "GetPRKeyWithPaperKey", []string { paperKey, userID })
                if err != nil {
                    return userType, errors.New("AccessControl"+ " : " + err.Error())
                }

                reviewerKey := string(reviewerKeyByte)

                err = CheckReviewerStatus(stub, reviewerKey, "accept")
                if err != nil {
                    return userType, errors.New("AccessControl"+ " : " + err.Error())
                }

                if paper.Status != "under_review" {
                    return userType, errors.New("AccessControl"+ " : " + paperKey + " is not in under_review statis.")
                }

                userType = ROP
            } else {
                acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                if err != nil {
                    return userType, errors.New("AccessControl"+ " : " + err.Error())
                }

                if OCCheck(paper, acceptanceModel) {
                    userType = REVIEWER
                } else {
                    return userType, errors.New("AccessControl"+ " : " + "OC error")
                }
            }

            return userType, nil
        //Author or Editor of Paper or broker
        case "AddRevisionNote" :
            commentByte, err := stub.GetState(args[0])
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }
            if len(commentByte) == 0 {
                return userType, errors.New("AccessControl" + " : " + "There is no " + args[0] + " comment.")
            }

            paperKey, err = CreatePaperKeyWithCommentKey(stub, args[0])
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            paper, err = GetPaper(stub, paperKey)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            err = CertifyAuthorWithPaper(stub, paper)
            if err == nil {
                userType = AUTHOR
            } else {
                err = CertifyEOPWithPaper(stub, paper)
                if err == nil {
                    userType = EOP
                } else {
                    err = CertifyBroker(stub)
                    if err == nil {
                        userType = BROKER
                    } else {
                        return userType, errors.New("AccessControl"+ " : " + "Ceritfy Error")
                    }
                }
            }

            return userType, nil
        //reviewer of paper or author or editor of paper
        // args[0] = commentKey or revisionNoteKey
        case "QueryComment", "QueryRevisionNote" :
            if function == "QueryComment" {
                commentKey = args[0]
            } else if function == "QueryRevisionNote" {
                commentKey, err = CreateCommentKeyWithRevisionNoteKey(stub, args[0])
                if err != nil {
                    return userType, errors.New("AccessControl" + " : " + err.Error())
                }
            }

            paperKey, err = CreatePaperKeyWithCommentKey(stub, commentKey)
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
                err = CertifyAuthorWithPaper(stub, paper)
                if err == nil {
                    userType = AUTHOR
                } else {
                    err = CertifyROPWithPaperKey(stub, paperKey)
                    if err == nil {
                        userType = ROP

                        reviewerKeyByte, err := InvokeChaincode(stub, "reviewer", "GetPRKeyWithPaperKey", []string { paperKey, userID })
                        if err != nil {
                            return userType, errors.New("AccessControl" + " : " + err.Error())
                        }

                        reviewerKey := string(reviewerKeyByte)
                        userIndex, err := GetReviewerIndex(stub, reviewerKey)
                        if err != nil {
                            return userType, errors.New("AccessControl" + " : " + err.Error())
                        }

                        targetIndex, err := GetReviewerIndex(stub, commentKey)
                        if err != nil {
                            return userType, errors.New("AccessControl" + " : " + err.Error())
                        }

                        if userIndex == targetIndex {
                            userType = SELF
                        } else {
                            roundString, err := GetRoundFromCommentKey(stub, commentKey)
                            if err != nil {
                                return userType, errors.New("AccessControl" + " : " + err.Error())
                            }

                            targetRound, err := strconv.Atoi(roundString)
                            if err != nil {
                                return userType, errors.New("AccessControl" + " : " + err.Error())
                            }

                            if targetRound > paper.Round {
                                return userType, errors.New("AccessControl" + " : " + "The comment does not exist.")
                            } else if targetRound == paper.Round {
                                userCommentKey, err := ConvertCommentKey(stub, commentKey, userID)
                                if err != nil {
                                    return userType, errors.New("AccessControl" + " : " + err.Error())
                                }

                                commentByte, err := stub.GetState(userCommentKey)
                                if err != nil {
                                    return userType, errors.New("AccessControl" + " : " + err.Error())
                                }
                                if len(commentByte) == 0 {
                                    return userType, errors.New("AccessControl" + " : " + "The comment does not exist.")
                                }
                            }
                        }
                    } else {
                        roundString, err := GetRoundFromCommentKey(stub, commentKey)
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

                        if (targetRound < paper.Round && (acceptanceModel.OR == "4" || acceptanceModel.OR == "5")) || ORCheck(paper, acceptanceModel) {
                            userType = REVIEWER
                        } else {
                            return userType, errors.New("AccessControl" + " : " + userID + " can't query " + commentKey + " comment.")
                        }
                    }
                }
            }

            return userType, nil
        //author or editor of paper
        // args[0] = contractKey
        case "QueryCommentWithContractKey", "QueryRevisionNoteWithContractKey" :
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
                err = CertifyAuthorWithPaper(stub, paper)
                if err == nil {
                    userType = AUTHOR
                } else {
                    err = CertifyROPWithPaperKey(stub, paperKey)
                    if err == nil {
                        userType = ROP
                    } else {
                        targetRoundString, err := GetRoundFromContractKey(stub, args[0])
                        if err != nil {
                          return userType, errors.New("AccessControl" + " : " + err.Error())
                        }

                        targetRound, err := strconv.Atoi(targetRoundString)
                        if err != nil {
                          return userType, errors.New("AccessControl" + " : " + err.Error())
                        }

                        acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                        if err != nil {
                          return userType, errors.New("AccessControl"+ " : " + err.Error())
                        }

                        if (targetRound < paper.Round && (acceptanceModel.OR == "4" || acceptanceModel.OR == "5")) || ORCheck(paper, acceptanceModel) {
                          userType = REVIEWER
                        } else {
                          return AERROR, nil
                        }
                    }
                }
            }

            return userType, nil
        //reviewer himself
        // args[0] = reviewerID
        case "QueryCommentWithReviewerID", "QueryRevisionNoteWithReviewerID" :
            err = CertifyIdentifier(stub, args[0])
            if err == nil {
                userType = SELF
            } else {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            return userType, nil
        //reviewer of paper or author or editor of paper
        // args[0] = paperKey
        case "QueryCommentWithPaperKey", "QueryRevisionNoteWithPaperKey" :
            paperKey = args[0]
            paper, err = GetPaper(stub, paperKey)
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            err = CertifyEOPWithPaper(stub, paper)
            if err == nil {
                userType = EOP
            } else {
                err = CertifyAuthorWithPaper(stub, paper)
                if err == nil {
                    userType = AUTHOR
                } else {
                    err = CertifyROPWithPaperKey(stub, paperKey)
                    if err == nil {
                        userType = ROP
                    } else {
                        targetRoundString, err := GetRoundFromContractKey(stub, args[0])
                        if err != nil {
                          return userType, errors.New("AccessControl" + " : " + err.Error())
                        }

                        targetRound, err := strconv.Atoi(targetRoundString)
                        if err != nil {
                          return userType, errors.New("AccessControl" + " : " + err.Error())
                        }

                        acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                        if err != nil {
                          return userType, errors.New("AccessControl"+ " : " + err.Error())
                        }

                        if (targetRound < paper.Round && (acceptanceModel.OR == "4" || acceptanceModel.OR == "5")) || ORCheck(paper, acceptanceModel) {
                          userType = REVIEWER
                        } else {
                          return AERROR, nil
                        }
                    }
                }
            }

            return userType, nil

        // editor of paper or author
        // args[0] = paperKey
        case "DeleteCommentWithPaperKey" :
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

func ApplyAcceptanceModel(stub shim.ChaincodeStubInterface, function string, userType UserType, target []byte) ([]byte, error) {
    var userID string
    var comment Comment
    var commentByte []byte
    var commentArray []Comment
    var paperKey string
    var paper *Paper
    var acceptanceModel *AcceptanceModel
    var buff bytes.Buffer
    var result []byte
    var writeOne bool
    var err error

    result = target
    writeOne = false

    switch function {
        // Comment
        case "QueryComment", "QueryRevisionNote" :
            if len(target) == 0 {
                return nil, errors.New("ApplyAcceptanceModel" + " : " + "The comment does not exist.")
            }

            if userType == EOP || userType == SELF {
                return target, nil
            } else {
                err = json.Unmarshal(target, &comment)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                paperKey, err = CreatePaperKeyWithContractKey(stub, comment.ContractKey)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                paper, err = GetPaper(stub, paperKey)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)

                if !OICheck(paper, acceptanceModel) {
                    comment.ReviewerID = ""
                }

                result, err = json.Marshal(comment)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                return result, nil
            }

        // Comment Array
        case "QueryCommentWithContractKey" :
            if userType == EOP {
                return target, nil
            } else {
                err = json.Unmarshal(target, &commentArray)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }
                if len(commentArray) == 0 {
                    return target, nil
                }

                paperKey, err = CreatePaperKeyWithContractKey(stub, commentArray[0].Key)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                paper, err = GetPaper(stub, paperKey)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                if userType == AUTHOR {
                    if OICheck(paper, acceptanceModel) {
                        return target, nil
                    } else {
                        buff.WriteString("[")

                        for _, comment := range commentArray {
                            comment.ReviewerID = ""
                            commentByte, err = json.Marshal(comment)
                            if err != nil {
                                return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                            }

                            if writeOne {
                                buff.WriteString(",")
                            }

                            writeOne = true
                            buff.Write(commentByte)
                        }

                        buff.WriteString("]")

                        return buff.Bytes(), nil
                    }
                } else if userType == ROP {
                    userID, err = GetIdentifier(stub)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    roundString, err := GetRoundFromCommentKey(stub, commentArray[0].Key)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    round, err := strconv.Atoi(roundString)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }
                    if round < paper.Round {
                        if OICheck(paper, acceptanceModel) {
                            return target, nil
                        } else {
                            buff.WriteString("[")

                            for _, comment := range commentArray {
                                if comment.ReviewerID != userID {
                                    comment.ReviewerID = ""
                                }

                                commentByte, err = json.Marshal(comment)
                                if err != nil {
                                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                                }

                                if writeOne {
                                    buff.WriteString(",")
                                }

                                writeOne = true
                                buff.Write(commentByte)
                            }

                            buff.WriteString("]")
                        }
                        return buff.Bytes(), nil

                    } else if round == paper.Round {
                        open := OICheck(paper, acceptanceModel)
                        isCommitted := false

                        buff.WriteString("[")

                        for _, comment := range commentArray {
                            if comment.ReviewerID == userID {
                                isCommitted = true
                            } else {
                                comment.ReviewerID = ""
                            }

                            commentByte, err = json.Marshal(comment)
                            if err != nil {
                                return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                            }

                            if writeOne {
                                buff.WriteString(",")
                            }

                            writeOne = true
                            buff.Write(commentByte)
                        }

                        buff.WriteString("]")

                        if !isCommitted {
                            return []byte("[]"), nil
                        } else if open {
                            return target, nil
                        }

                        return buff.Bytes(), nil
                    } else {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + "There is a higher round(" + roundString + ") of comments than the round of papers.")
                    }
                } else if userType == REVIEWER {
                    if OICheck(paper, acceptanceModel) {
                      return target, nil
                    } else {
                      buff.WriteString("[")

                        for _, comment := range commentArray {
                          comment.ReviewerID = ""

                          commentByte, err = json.Marshal(comment)
                            if err != nil {
                              return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                            }

                          if writeOne {
                            buff.WriteString(",")
                          }

                          writeOne = true
                            buff.Write(commentByte)
                        }

                      buff.WriteString("]")
                    }
                    return buff.Bytes(), nil

                } else {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + "Certify Error")
                }
            }

        case "QueryCommentWithReviewerID", "QueryRevisionNoteWithRevisionNote" :
            return target, nil

        case "QueryCommentWithPaperKey", "QueryRevisionNoteWithPaperKey" :
            if userType == EOP {
                return target, nil
            } else {
                err = json.Unmarshal(target, &commentArray)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }
                if len(commentArray) == 0 {
                    return target, nil
                }

                paperKey, err = CreatePaperKeyWithContractKey(stub, commentArray[0].Key)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                paper, err = GetPaper(stub, paperKey)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                if userType == AUTHOR {
                    if OICheck(paper, acceptanceModel) {
                        return target, nil
                    } else {
                        buff.WriteString("[")

                        for _, comment := range commentArray {
                            comment.ReviewerID = ""
                            commentByte, err = json.Marshal(comment)
                            if err != nil {
                                return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                            }

                            if writeOne {
                                buff.WriteString(",")
                            }

                            writeOne = true
                            buff.Write(commentByte)
                        }

                        buff.WriteString("]")

                        return buff.Bytes(), nil
                    }
                } else if userType == ROP {
                    userID, err = GetIdentifier(stub)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    roundString, err := GetRoundFromCommentKey(stub, commentArray[0].Key)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    round, err := strconv.Atoi(roundString)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    open := OICheck(paper, acceptanceModel)
                    isCommitted := false

                    buff.WriteString("[")

                    for _, comment := range commentArray {
                        if comment.ReviewerID == userID && round == paper.Round {
                            isCommitted = true
                            break
                        }
                    }

                    for _, comment := range commentArray {
                        commentRoundString, err := GetRoundFromCommentKey(stub, comment.Key)
                        if err != nil {
                            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                        }

                        commentRound, err := strconv.Atoi(commentRoundString)
                        if err != nil {
                            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                        }
                        if !isCommitted && commentRound == paper.Round {
                            continue
                        } else {
                            if !open && comment.ReviewerID != userID {
                                comment.ReviewerID = ""
                            }

                            commentByte, err = json.Marshal(comment)
                            if err != nil {
                                return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                            }

                            if writeOne {
                                buff.WriteString(",")
                            }

                            writeOne = true
                            buff.Write(commentByte)
                        }
                    }

                    buff.WriteString("]")

                    return buff.Bytes(), nil
                } else {
                    open := OICheck(paper, acceptanceModel)

                    buff.WriteString("[")

                    for _, comment := range commentArray {
                        commentRoundString, err := GetRoundFromCommentKey(stub, comment.Key)
                        if err != nil {
                            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                        }

                        commentRound, err := strconv.Atoi(commentRoundString)
                        if err != nil {
                            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                        }

                        if commentRound >= paper.Round {
                            continue
                        } else {
                            if !open {
                                comment.ReviewerID = ""
                            }

                            commentByte, err = json.Marshal(comment)
                            if err != nil {
                                return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                            }

                            if writeOne {
                                buff.WriteString(",")
                            }

                            writeOne = true
                            buff.Write(commentByte)
                        }
                    }

                    buff.WriteString("]")

                    return buff.Bytes(), nil
                }
            }

        default :
            return nil, errors.New("ApplyAcceptanceModel" + " : " + function + " is Invalid Smart Contract function name.")
    }
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
        return shim.Error("comment : " + err.Error())
    }

    switch function {
        // No Return
        case "AddComment" :
            err = s.AddComment(stub, args)
        case "AddRevisionNote" :
            err = s.AddRevisionNote(stub, args)
        case "DeleteCommentWithPaperKey" :
            err = s.DeleteCommentWithPaperKey(stub, args)

        // Comment
        case "QueryComment" :
            result, err = s.QueryComment(stub, args)
        case "QueryRevisionNote" :
            result, err = s.QueryRevisionNote(stub, args)

        // Comment Array
        case "QueryCommentWithContractKey" :
            if userType == AERROR {
                return shim.Success([]byte("[]"))
            }
            result, err = s.QueryCommentWithContractKey(stub, args)
        case "QueryCommentWithReviewerID" :
            if userType == AERROR {
                return shim.Success([]byte("[]"))
            }
            result, err = s.QueryCommentWithReviewerID(stub, args)
        case "QueryCommentWithPaperKey" :
            if userType == AERROR {
                return shim.Success([]byte("[]"))
            }
            result, err = s.QueryCommentWithPaperKey(stub, args)
        case "QueryRevisionNoteWithContractKey" :
            if userType == AERROR {
                return shim.Success([]byte("[]"))
            }
            result, err = s.QueryRevisionNoteWithContractKey(stub, args)
        case "QueryRevisionNoteWithReviewerID" :
            if userType == AERROR {
                return shim.Success([]byte("[]"))
            }
            result, err = s.QueryRevisionNoteWithReviewerID(stub, args)
        case "QueryRevisionNoteWithPaperKey" :
            if userType == AERROR {
                return shim.Success([]byte("[]"))
            }
            result, err = s.QueryRevisionNoteWithPaperKey(stub, args)
    }

    if err != nil {
        return shim.Error("comment : " + err.Error())
    }

    if result != nil {
        maskedResult, err = ApplyAcceptanceModel(stub, function, userType, result)
        if err != nil {
            return shim.Error("comment : " + err.Error())
        }
    }

    return shim.Success(maskedResult)
}

//args = [ contractKey, location, comment ]
func (s *SmartContract) AddComment(stub shim.ChaincodeStubInterface, args []string) error {
    var comment Comment
    var commentByte []byte
    var identifier string
    var email string
    var CRKey string
    var RCKey string
    var paperKey string
    var location []string
    var message []string
    var RPReviewerKey string
    var reviewerKey string
    var reviewerIndex string
    var reviewer Reviewer
    var reviewerByte []byte
    var err error

    if len(args) != 3 {
        return errors.New("Incorrect number of arguments. Expecting 3")
    }

    paperKey, err = CreatePaperKeyWithContractKey(stub, args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    identifier, err = GetIdentifier(stub)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    RPReviewerKey, err = ConvertReviewerKey(stub, paperKey, identifier)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    reviewerIndex, err = GetReviewerIndex(stub, RPReviewerKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    reviewerKey, err = CreateReviewerKeyWithPaperKey(stub, paperKey, reviewerIndex)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    reviewerByte, err = InvokeChaincode(stub, "reviewer", "QueryReviewer", []string{ reviewerKey })
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = json.Unmarshal(reviewerByte, &reviewer)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    email = reviewer.Email

    CRKey, err = CreateCommentKeyWithContractKey(stub, args[0], reviewerIndex)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    RCKey, err = ConvertCommentKey(stub, CRKey, identifier)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = json.Unmarshal([]byte(args[1]), &location)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = json.Unmarshal([]byte(args[2]), &message)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    comment = Comment { Key : CRKey, ContractKey : args[0], ReviewerID : identifier, ReviewerEmail : email, Location : location, Comment : message }
    commentByte, err = json.Marshal(comment)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(CRKey, commentByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(RCKey, []byte(CRKey))
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    _, err = InvokeChaincode(stub, "reviewer", "UpdateStatus", []string { reviewerKey, "submitted" })
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

//args = [ commentKey, location, revisionNote]
func (s *SmartContract) AddRevisionNote(stub shim.ChaincodeStubInterface, args []string) error {
    var identifier string
    var email string
    var contractKey string
    var commentByte []byte
    var revisionKey string
    var location []string
    var message []string
    var comment Comment
    var err error

    if len(args) != 3 {
        return errors.New("Incorrect number of arguments. Expecting 3")
    }

    identifier, err = GetIdentifier(stub)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    email, err = GetEmail(stub)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    revisionKey, err = CreateRevisionNoteKeyWithCommentKey(stub, args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    contractKey, err = CreateContractKeyWithCommentKey(stub, args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = json.Unmarshal([]byte(args[1]), &location)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = json.Unmarshal([]byte(args[2]), &message)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    comment = Comment { Key : args[0], ContractKey : contractKey, ReviewerID : identifier, ReviewerEmail : email, Location : location, Comment : message }
    commentByte, err = json.Marshal(comment)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(revisionKey, commentByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

//args = [ paperKey ]
func (s *SmartContract) DeleteCommentWithPaperKey(stub shim.ChaincodeStubInterface, args []string) error {
    var commentIter shim.StateQueryIteratorInterface
    var splitPaperKey []string
    var err error

    if len(args) != 1 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    _, splitPaperKey, err = stub.SplitCompositeKey(args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    commentIter, err = stub.GetStateByPartialCompositeKey("CR", splitPaperKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if !commentIter.HasNext() {
        return nil
    }

    for commentIter.HasNext() {
        commentKV, err := commentIter.Next()
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        err = stub.DelState(commentKV.Key)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }
    }

    return nil
}

//args = [ commentKey ]
func (s *SmartContract) QueryComment(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var commentByte []byte
    var err error

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    commentByte, err = stub.GetState(args[0])
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if len(commentByte) == 0 {
        return nil, errors.New(function + " : " + "There are no comment have " + args[0] + " commentKey.")
    }

    return commentByte, nil
}

//args = [ paperKey ]
func (s *SmartContract) QueryCommentWithPaperKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var splitPaperKey []string
    var commentIter shim.StateQueryIteratorInterface
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

    commentIter, err = stub.GetStateByPartialCompositeKey("CR", splitPaperKey)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !commentIter.HasNext() {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for commentIter.HasNext() {
        commentKV, err := commentIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if flag {
            buff.WriteString(",")
        } else {
            flag = true
        }

        buff.Write(commentKV.Value)
    }

    commentIter.Close()
    buff.WriteString("]")

    return buff.Bytes(), nil
}

//args = [ contractKey ]
func (s *SmartContract) QueryCommentWithContractKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var splitContractKey []string
    var commentIter shim.StateQueryIteratorInterface
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

    commentIter, err = stub.GetStateByPartialCompositeKey("CR", splitContractKey)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !commentIter.HasNext() {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for commentIter.HasNext() {
        commentKV, err := commentIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if flag {
            buff.WriteString(",")
        } else {
            flag = true
        }

        buff.Write(commentKV.Value)
    }

    commentIter.Close()
    buff.WriteString("]")

    return buff.Bytes(), nil
}

//args = [ reviewerID ]
func (s *SmartContract) QueryCommentWithReviewerID(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var CRKeyIter shim.StateQueryIteratorInterface
    var buff bytes.Buffer
    var commentByte []byte
    var flag bool
    var err error

    flag = false

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    CRKeyIter, err = stub.GetStateByPartialCompositeKey("RC", []string { args[0] })
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !CRKeyIter.HasNext() {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for CRKeyIter.HasNext() {
        CRKeyKV, err := CRKeyIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        commentByte, err = stub.GetState(string(CRKeyKV.Value))
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }
        if len(commentByte) == 0 {
            return nil, errors.New(function + " : " + "There are no comments written by " + args[0] + ".")
        }

        if flag {
            buff.WriteString(",")
        } else {
            flag = true
        }

        buff.Write(commentByte)
    }

    CRKeyIter.Close()
    buff.WriteString("]")

    return buff.Bytes(), nil
}

//args = [ revisionNoteKey ]
func (s *SmartContract) QueryRevisionNote(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var revisionNoteByte []byte
    var err error

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    revisionNoteByte, err = stub.GetState(args[0])
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if len(revisionNoteByte) == 0 {
        return nil, errors.New(function + " : " + "There are no revisionNote have " + args[0] + " revisionNoteKey.")
    }

    return revisionNoteByte, nil
}

//args = [ paperKey ]
func (s *SmartContract) QueryRevisionNoteWithPaperKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var splitPaperKey []string
    var revisionNoteIter shim.StateQueryIteratorInterface
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

    revisionNoteIter, err = stub.GetStateByPartialCompositeKey("RN", splitPaperKey)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !revisionNoteIter.HasNext() {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for revisionNoteIter.HasNext() {
        revisionNoteKV, err := revisionNoteIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if flag {
            buff.WriteString(",")
        } else {
            flag = true
        }

        buff.Write(revisionNoteKV.Value)
    }

    revisionNoteIter.Close()
    buff.WriteString("]")

    return buff.Bytes(), nil
}

//args = [ contractKey ]
func (s *SmartContract) QueryRevisionNoteWithContractKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var splitContractKey []string
    var revisionNoteIter shim.StateQueryIteratorInterface
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

    revisionNoteIter, err = stub.GetStateByPartialCompositeKey("RN", splitContractKey)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !revisionNoteIter.HasNext() {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for revisionNoteIter.HasNext() {
        revisionNoteKV, err := revisionNoteIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if flag {
            buff.WriteString(",")
        } else {
            flag = true
        }

        buff.Write(revisionNoteKV.Value)
    }

    revisionNoteIter.Close()
    buff.WriteString("]")

    return buff.Bytes(), nil
}

//args = [ reviewerID ]
func (s *SmartContract) QueryRevisionNoteWithReviewerID(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var CRKeyIter shim.StateQueryIteratorInterface
    var revisionNoteKey string
    var buff bytes.Buffer
    var revisionNoteByte []byte
    var flag bool
    var err error

    flag = false

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    CRKeyIter, err = stub.GetStateByPartialCompositeKey("RC", []string { args[0] })
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !CRKeyIter.HasNext() {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for CRKeyIter.HasNext() {
        CRKeyKV, err := CRKeyIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        revisionNoteKey, err = CreateRevisionNoteKeyWithCommentKey(stub, string(CRKeyKV.Value))
        revisionNoteByte, err = stub.GetState(revisionNoteKey)
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }
        if len(revisionNoteByte) == 0 {
            return nil, errors.New(function + " : " + "There are no revisionNotes written by " + args[0] + ".")
        }

        if flag {
            buff.WriteString(",")
        } else {
            flag = true
        }

        buff.Write(revisionNoteByte)
    }

    CRKeyIter.Close()
    buff.WriteString("]")

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
