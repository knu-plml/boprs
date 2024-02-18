package main
import (
    "encoding/json"
    "fmt"
    "strconv"
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

func GetNumberOfReviewer(stub shim.ChaincodeStubInterface, paperKey string) (int, error) {
    var splitedPaperKey []string
    var reviewerIter shim.StateQueryIteratorInterface
    var reviewerCount int
    var err error

    _, splitedPaperKey, err = stub.SplitCompositeKey(paperKey)
    if err != nil {
        return -1, errors.New(function + " : " + err.Error())
    }

    reviewerIter, err = stub.GetStateByPartialCompositeKey("PR", splitedPaperKey)
    if err != nil {
        return -1, errors.New(function + " : " + err.Error())
    }

    reviewerCount = 0
    for reviewerIter.HasNext() {
        reviewerCount += 1
        _, err = reviewerIter.Next()
        if err != nil {
            return -1, errors.New(function + " : " + err.Error())
        }
    }

    return reviewerCount, nil
}

func AccessControl(stub shim.ChaincodeStubInterface, function string, args []string) (UserType, error) {
    var userType UserType
    var reviewerByte []byte
    var reviewer Reviewer
    var paperKey string
    var paper *Paper
    var err error

    userType = REVIEWER

    switch function {
        //Anyone
        //args[0] = paperKey
        case "QueryReviewerWithPaperKey" :
            paperKey = args[0]
            paper, err = GetPaper(stub, paperKey)
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            err = CertifyEOOWithPaper(stub, paper)
            if err == nil {
                userType = EOO
            } else {
                err = CertifyEOPWithPaper(stub, paper)
                if err == nil {
                    userType = EOP
                } else {
                    err = CertifyAuthorWithPaper(stub, paper)
                    if err == nil {
                        userType = AUTHOR
                    } else {
					    identifier, err := GetIdentifier(stub)
					    if err != nil {
					        return userType, errors.New("AccessControl : " + err.Error())
					    }

					    RPKey, err := ConvertReviewerKey(stub, paperKey, identifier)
					    if err != nil {
					        return userType, errors.New("AccessControl : " + err.Error())
					    }

					    reviewerKeyByte, err := stub.GetState(RPKey)
					    if err != nil {
					        return userType, errors.New("AccessControl : " + err.Error())
					    }
					    if len(reviewerKeyByte) == 0 {
					        return userType, errors.New("AccessControl : " + identifier + " is not reviewer of " + paperKey + " paper.")
					    }

					    reviewerKey := string(reviewerKeyByte)

					    reviewerByte, err = stub.GetState(reviewerKey)
					    if err != nil {
					        return userType, errors.New("AccessControl : " + err.Error())
					    }
					    if len(reviewerKeyByte) == 0 {
					        return userType, errors.New("AccessControl : " + identifier + " is not reviewer of " + paperKey + " paper.")
					    }

					    err = json.Unmarshal(reviewerByte, &reviewer)
					    if err != nil {
					        return userType, errors.New("AccessControl : " + err.Error())
					    }
					    if reviewer.Status != "selected" && reviewer.Status != "accept" && reviewer.Status != "submitted" {
					        return userType, errors.New("AccessControl : " + "CertifyError")
					    } else {
                          userType = ROP
                        }
                    }
                }
            }

            return userType, nil
        // editor of paper
        //args[0] = paperKey
        case "QueryCandidateIDWithPaperKey", "QueryReviewerToRate" :
            err = CertifyEOPWithPaperKey(stub, args[0])
            if err == nil {
                userType = EOP
            } else {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }


            return userType, nil
        //Anyone
        //args[0] = reviewerKey
        case "QueryReviewer" :
            paperKey, err = CreatePaperKeyWithReviewerKey(stub, args[0])
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            paper, err = GetPaper(stub, paperKey)
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            err = CertifyEOOWithPaper(stub, paper)
            if err == nil {
                userType = EOO
            } else {
                err = CertifyEOPWithPaper(stub, paper)
                if err == nil {
                    userType = EOP
                } else {
                    err = CertifyAuthorWithPaper(stub, paper)
                    if err == nil {
                        userType = AUTHOR
                    } else {
                        reviewerByte, err = stub.GetState(args[0])
                        if err != nil {
                            return userType, errors.New("AccessControl"+ " : " + err.Error())
                        }

                        err = json.Unmarshal(reviewerByte, &reviewer)
                        if err != nil {
                            return userType, errors.New("AccessControl"+ " : " + err.Error())
                        }

                        err = CertifyIdentifier(stub, reviewer.ReviewerID)
                        if err == nil {
                            userType = SELF
                        } else {
                            return userType, errors.New("AccessControl"+ " : " + err.Error())
                        }
                    }
                }
            }

            return userType, nil
        //editor or reviewer himself 
        //args[0] = reviewerID
        case "QueryReviewerWithReviewerID", "QueryReviewPapers", "UpdateEmail" :
            err = CertifyEditor(stub)
            if err == nil {
                userType = EDITOR
            } else {
                err = CertifyIdentifier(stub, args[0])
                if err == nil {
                    userType = SELF
                } else {
                    return userType, errors.New("AccessControl"+ " : " + err.Error())
                }
            }

            return userType, nil

        // Reviewer himself or Editor of Paper 
        // args[0] = RPKey
        case "GetPRKey" :
            _, splitedRPKey, err := stub.SplitCompositeKey(args[0])
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            userID, err := GetIdentifier(stub)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            if userID == splitedRPKey[0] {
                return SELF, nil
            }

            paperKey, err := CreatePaperKey(stub, splitedRPKey[1], splitedRPKey[2])
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            err = CertifyEOPWithPaperKey(stub, paperKey)
            if err == nil {
                return EOP, nil
            }

            return userType, nil

        // Reviewer himself or Editor of Paper 
        // args[0] = paperKey args[1] = reviewerID
        case "GetPRKeyWithPaperKey" :
            userID, err := GetIdentifier(stub)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            if userID == args[1] {
                return SELF, nil
            }

            err = CertifyEOPWithPaperKey(stub, args[0])
            if err == nil {
                return EOP, nil
            }

            return userType, nil

        // Anyone 
        //args[0] = paperKey
        case "RegisterCandidate" :
            userType = SELF
            paper, err = GetPaper(stub, args[0])
            if paper.Status !=  "recruit_reviewer" {
                return userType, errors.New("AccessControl"+ " : " + args[0] + " paper does not recruit reviewers.")
            }

            acceptanceModel, err := GetAcceptanceModel(stub, paper.Organization)
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            if acceptanceModel.OP == "0" {
                organization, err := GetMSPID(stub)
                if err != nil {
                    return userType, errors.New("AccessControl"+ " : " + err.Error())
                }

                if organization != paper.Organization {
                    return userType, errors.New("AccessControl"+ " : " + "By AcceptanceModel, reviewers belonging to " + organization + " cannot participate in reviews belonging to " + paper.Organization + ".")
                }
            }

            return userType, nil
        // reviewer himself 
        //args[0] = paperKey
        case "DeclineReview" :
            paperKey = args[0]
			identifier, err := GetIdentifier(stub)
			if err != nil {
			    return userType, errors.New("AccessControl : " + err.Error())
			}

			RPKey, err := ConvertReviewerKey(stub, paperKey, identifier)
			if err != nil {
			    return userType, errors.New("AccessControl : " + err.Error())
			}

			reviewerKeyByte, err := stub.GetState(RPKey)
			if err != nil {
			    return userType, errors.New("AccessControl : " + err.Error())
			}
			if len(reviewerKeyByte) == 0 {
			    return userType, errors.New("AccessControl : " + identifier + " is not reviewer of " + paperKey + " paper.")
			}

			reviewerKey := string(reviewerKeyByte)

			reviewerByte, err = stub.GetState(reviewerKey)
			if err != nil {
			    return userType, errors.New("AccessControl : " + err.Error())
			}
			if len(reviewerKeyByte) == 0 {
			    return userType, errors.New("AccessControl : " + identifier + " is not reviewer of " + paperKey + " paper.")
			}

			err = json.Unmarshal(reviewerByte, &reviewer)
			if err != nil {
			    return userType, errors.New("AccessControl : " + err.Error())
			}
			if reviewer.Status != "selected" && reviewer.Status != "accept" && reviewer.Status != "submitted" {
			    return userType, errors.New("AccessControl : " + "CertifyError")
			} else {
              userType = SELF
            }

            return userType, nil
        // editor of paper 
        //args[0] = PRKey
        case "RejectReviewer" :
            paperKey, err = CreatePaperKeyWithReviewerKey(stub, args[0])
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            err = CertifyEOPWithPaperKey(stub, paperKey)
            if err == nil {
                userType = EOP
            } else {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            return userType, nil
        //broker
        //args[0] = paperKey
        case "RegisterReviewer" :
            paperKey = args[0]

            err = CertifyBroker(stub)
            if err == nil {
                userType = EOP
            } else {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            err = CheckPaperStatus(stub, paperKey, "reviewer_invited")
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            return userType, nil
        //editor of paper 
        //args[0] = paperKey
        case "SelectReviewer", "StopRecruiting" :
            paperKey = args[0]

            err = CheckPaperStatus(stub, paperKey, "recruit_reviewer")
            if err != nil {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            err = CertifyEOPWithPaperKey(stub, paperKey)
            if err == nil {
                userType = EOP
            } else {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            return userType, nil

        // editor of paper or reviewer himself
        //args[0] = PRKey
        case "UpdateStatus" :
            paperKey, err = CreatePaperKeyWithReviewerKey(stub, args[0])
            if err != nil {
                return userType, errors.New("AccessControl" + " : " + err.Error())
            }

            err = CertifyEOPWithPaperKey(stub, paperKey)
            if err == nil {
                userType = EOP
            } else {
                reviewerByte, err = stub.GetState(args[0])
                if err != nil {
                    return userType, errors.New("AccessControl"+ " : " + err.Error())
                }

                err = json.Unmarshal(reviewerByte, &reviewer)
                if err != nil {
                    return userType, errors.New("AccessControl"+ " : " + err.Error())
                }

                err = CertifyIdentifier(stub, reviewer.ReviewerID)
                if err == nil {
                    userType = SELF
                } else {
                    return userType, errors.New("AccessControl"+ " : " + err.Error())
                }
            }

            return userType, nil

        //editor of paper 
        //args[0] = paperKey
        case "InitReviewer" :
            paperKey = args[0]

            err = CertifyEOPWithPaperKey(stub, paperKey)
            if err == nil {
                userType = EOP
            } else {
                return userType, errors.New("AccessControl"+ " : " + err.Error())
            }

            return userType, nil

        //editor of paper of author
        //args[0] = paperKey
        case "DeleteReviewerWithPaperKey" :
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
    var paper *Paper
    var reviewer Reviewer
    var reviewerByte []byte
    var reviewerArray []Reviewer
    var paperArray []Paper
    var paperByte []byte
    var acceptanceModel *AcceptanceModel
    var buff bytes.Buffer
    var result []byte
    var writeOne bool
    var err error


    result = target
    writeOne = false

    switch function {
        //return Reviewer
        case "QueryReviewer" :
            if userType == EOP || userType == EOO || userType == SELF {
                return target, nil
            } else {
                err = json.Unmarshal(target, &reviewer)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                paper, err = GetPaper(stub, reviewer.PaperKey)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                if !OICheck(paper, acceptanceModel) {
                    reviewer.ReviewerID = ""
                    reviewer.Email = ""
                    result, err = json.Marshal(reviewer)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }
                }

                return result, nil
            }

        // reutn Reviewer Array
        case "QueryCandidateIDWithPaperKey", "QueryReviewerToRate" :
            if userType == EOP {
                return target, nil
            } else {
                return nil, errors.New("ApplyAcceptanceModel" + " : " + "accessControle Error, this user is not editor of paper")
            }

        case "QueryReviewerWithPaperKey" :
            if userType == EOP || userType == EOO {
                return target, nil
            } else {
                err = json.Unmarshal(target, &reviewerArray)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                if len(reviewerArray) == 0 {
                    return target, nil
                }

                paper, err = GetPaper(stub, reviewerArray[0].PaperKey)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                open := OICheck(paper, acceptanceModel)
                buff.WriteString("[")

                for _, reviewer := range reviewerArray {
                    err = CertifyIdentifier(stub, reviewer.ReviewerID)
                    if err != nil && !open {
                        reviewer.ReviewerID = ""
                        reviewer.Email = ""
                        result, err = json.Marshal(reviewer)
                        if err != nil {
                            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                        }
                    }

                    reviewerByte, err = json.Marshal(reviewer)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    if writeOne {
                        buff.WriteString(",")
                    }

                    writeOne = true
                    buff.Write(reviewerByte)
                }

                buff.WriteString("]")

                return buff.Bytes(), nil
            }

        case "QueryReviewerWithReviewerID" :
            if userType == SELF {
                return target, nil
            } else if userType == EDITOR {
                err = json.Unmarshal(target, &reviewerArray)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }

                buff.WriteString("[")

                for _, reviewer := range reviewerArray {
                    paper, err = GetPaper(stub, reviewer.PaperKey)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    err = CertifyEOOWithPaper(stub, paper)
                    if err == nil || OICheck(paper, acceptanceModel) {
                        reviewerByte, err = json.Marshal(reviewer)
                        if err != nil {
                            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                        }

                        if writeOne {
                            buff.WriteString(",")
                        }

                        writeOne = true
                        buff.Write(reviewerByte)
                    }
                }

                buff.WriteString("]")

                return buff.Bytes(), nil
            } else {
                err = json.Unmarshal(target, &reviewerArray)
                if err != nil {
                    return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                }
                if len(reviewerArray) == 0 {
                    return target, nil
                }

                buff.WriteString("[")

                for _, reviewer := range reviewerArray {
                    paper, err = GetPaper(stub, reviewer.PaperKey)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    if OICheck(paper, acceptanceModel) {
                        reviewerByte, err = json.Marshal(reviewer)
                        if err != nil {
                            return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                        }

                        if writeOne {
                            buff.WriteString(",")
                        }

                        writeOne = true
                        buff.Write(reviewerByte)
                    }
                }

                buff.WriteString("]")

                return buff.Bytes(), nil
            }

        // return Paper Array
        case "QueryReviewPapers" :
            err = json.Unmarshal(target, &paperArray)
            if err != nil {
                return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
            }
            if len(paperArray) == 0 {
                return target, nil
            }

            if userType == SELF {
                buff.WriteString("[")

                for _, paper := range paperArray {
                    acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    if !OICheck(&paper, acceptanceModel) {
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

            } else if userType == EDITOR {
                buff.WriteString("[")

                for _, paper := range paperArray {
                    acceptanceModel, err = GetAcceptanceModel(stub, paper.Organization)
                    if err != nil {
                        return nil, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
                    }

                    err = CertifyEOOWithPaper(stub, &paper)
                    if err == nil || OICheck(&paper, acceptanceModel) {
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
                return nil, errors.New("ApplyAcceptanceModel" + " : " + "Certify Error")
            }

        case "GetPRKey", "GetPRKeyWithPaperKey" :
            if userType == SELF || userType == EOP {
                return target, nil
            } else {
                return nil, errors.New("ApplyAcceptanceModel" + " : " + "accessControle Error")
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
        return shim.Error("reviewer : " + err.Error())
    }

    switch function {
        //return Reviewer
        case "QueryReviewer" :
            result, err = s.QueryReviewer(stub, args)

        // result, err = Reviewer Array
        case "QueryCandidateIDWithPaperKey" :
            result, err = s.QueryCandidateIDWithPaperKey(stub, args)
        case "QueryReviewerWithPaperKey" :
            result, err = s.QueryReviewerWithPaperKey(stub, args)
        case "QueryReviewerWithReviewerID" :
            result, err = s.QueryReviewerWithReviewerID(stub, args)
//        case "QueryReviewerToRate" :
//            result, err = s.QueryReviewerToRate(stub, args)

        // result, err = Paper Array
        case "QueryReviewPapers" :
            result, err = s.QueryReviewPapers(stub, args)

        // result, err = PRKey
        case "GetPRKey" :
            result, err = s.GetPRKey(stub, args)
        case "GetPRKeyWithPaperKey" :
            result, err = s.GetPRKeyWithPaperKey(stub, args)

        //NO return
        case "RegisterCandidate" :
            err = s.RegisterCandidate(stub, args)
        case "RegisterReviewer" :
            err = s.RegisterReviewer(stub, args)
        case "StopRecruiting" :
            err = s.StopRecruiting(stub, args)
        case "DeclineReview" :
            err = s.DeclineReview(stub, args)
        case "RejectReviewer" :
            err = s.RejectReviewer(stub, args)
        case "SelectReviewer" :
            err = s.SelectReviewer(stub, args)
        case "UpdateStatus" :
            err = s.UpdateStatus(stub, args)
        case "UpdateEmail" :
            err = s.UpdateEmail(stub, args)
        case "InitReviewer" :
            err = s.InitReviewer(stub, args)
        case "DeleteReviewerWithPaperKey" :
            err = s.DeleteReviewerWithPaperKey(stub, args)
    }

    if err != nil {
        return shim.Error("reviewer : " + err.Error())
    }

    //maskedResult = result
    if result != nil {
        maskedResult, err = ApplyAcceptanceModel(stub, function, userType, result)
        if err != nil {
            return shim.Error("reviewer : " + err.Error())
        }
    }

    return shim.Success(maskedResult)
}

//args = [ PaperKey ]
func (s *SmartContract) QueryReviewerWithPaperKey (stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var reviewerIter shim.StateQueryIteratorInterface
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

    reviewerIter, err = stub.GetStateByPartialCompositeKey("PR", splitPaperKey )
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if(!reviewerIter.HasNext()) {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for reviewerIter.HasNext() {
        reviewerKV, err := reviewerIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if flag {
            buff.WriteString(",")
        } else {
            flag = true
        }

        buff.Write(reviewerKV.Value)
    }

    buff.WriteString("]")
    reviewerIter.Close()

    return buff.Bytes(), nil
}

// args = [ paperKey ]
func (s *SmartContract) QueryCandidateIDWithPaperKey (stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var reviewerIter shim.StateQueryIteratorInterface
    var splitPaperKey []string
    var reviewer Reviewer
    var buff bytes.Buffer
    var flag bool
    var err error

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    _, splitPaperKey, err = stub.SplitCompositeKey(args[0])
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }

    reviewerIter, err = stub.GetStateByPartialCompositeKey("PR", splitPaperKey )
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }

    flag = false
    buff.WriteString("[")
    for reviewerIter.HasNext() {
        reviewerKV, err := reviewerIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        err = json.Unmarshal(reviewerKV.Value, &reviewer)
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if reviewer.Status == "candidate" {
            if flag {
                buff.WriteString(",")
            } else {
                flag = true
            }

            buff.WriteString("{\"ORCID\":\"")
            buff.Write([]byte(reviewer.ReviewerID))
            buff.WriteString("\",\"email\":\"")
            buff.Write([]byte(reviewer.Email))
            buff.WriteString("\"}")
        }
    }

    buff.WriteString("]")
    reviewerIter.Close()

    return buff.Bytes(), nil
}

//args = [ reviewerID ]
func (s *SmartContract) QueryReviewPapers (stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var PRKeyIter shim.StateQueryIteratorInterface
    var reviewerByte []byte
    var reviewer Reviewer
    var paperByte []byte
    var buff bytes.Buffer
    var flag bool
    var err error

    flag = false

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    PRKeyIter, err = stub.GetStateByPartialCompositeKey("RP", []string{ args[0] })
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if(!PRKeyIter.HasNext()) {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for PRKeyIter.HasNext() {
        RPKeyKV, err := PRKeyIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        reviewerByte, err = stub.GetState(string(RPKeyKV.Value))
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        err = json.Unmarshal(reviewerByte, &reviewer)
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        paperByte, err = GetPaperByte(stub, reviewer.PaperKey)
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
    PRKeyIter.Close()

    return buff.Bytes(), nil
}

//args = [ reviewerID ]
func (s *SmartContract) QueryReviewerWithReviewerID (stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var PRKeyIter shim.StateQueryIteratorInterface
    var reviewerByte []byte
    var buff bytes.Buffer
    var flag bool
    var err error

    flag = false

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    PRKeyIter, err = stub.GetStateByPartialCompositeKey("RP", []string{ args[0] })
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !PRKeyIter.HasNext() {
        return []byte("[]"), nil
    }

    buff.WriteString("[")

    for PRKeyIter.HasNext() {
        PRKeyKV, err := PRKeyIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        reviewerByte, err = stub.GetState(string(PRKeyKV.Value))
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if flag {
            buff.WriteString(",")
        } else {
            flag = true
        }

        buff.Write(reviewerByte)
    }

    buff.WriteString("]")
    PRKeyIter.Close()

    return buff.Bytes(), nil
}

//args = [ reviewerKey ]
func (s *SmartContract) QueryReviewer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var reviewerByte []byte
    var err error

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    reviewerByte, err = stub.GetState(args[0])
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if len(reviewerByte) == 0 {
        return nil, errors.New(function + " : " + args[0] + " is not reviewer.")
    }

    return reviewerByte, nil
}

//args = [ RPKey ]
func (s *SmartContract) GetPRKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var PRKeyByte []byte
    var err error

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    PRKeyByte, err = stub.GetState(args[0])
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if len(PRKeyByte) == 0 {
        return nil, errors.New(function + " : " + args[0] + " is not reviewer.")
    }

    return PRKeyByte, nil
}

//args = [ paperKey, reviewerID ]
func (s *SmartContract) GetPRKeyWithPaperKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var RPKey string
    var PRKeyByte []byte
    var err error

    if len(args) != 2 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 2")
    }

    RPKey, err = ConvertReviewerKey(stub, args[0], args[1])
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }

    PRKeyByte, err = stub.GetState(RPKey)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if len(PRKeyByte) == 0 {
        return nil, errors.New(function + " : " + args[0] + " is not reviewer.")
    }

    return PRKeyByte, nil
}

//args = [ paperKey ]
func (s *SmartContract) RegisterCandidate(stub shim.ChaincodeStubInterface, args []string) error {
    var reviewerCount int
    var reviewerByte []byte
    var reviewer Reviewer
    var identifier string
    var email string
    var PRKey string
    var PRKeyByte []byte
    var RPKey string
    var err error

    if len(args) != 1 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    identifier, err = GetIdentifier(stub)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    email, err = GetEmail(stub)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    reviewerCount, err = GetNumberOfReviewer(stub, args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    PRKey, err = CreateReviewerKeyWithPaperKey(stub, args[0], strconv.Itoa(reviewerCount + 1))
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    RPKey, err = ConvertReviewerKey(stub, PRKey, identifier)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    PRKeyByte, err = stub.GetState(RPKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if PRKeyByte != nil {
        PRKey = string(PRKeyByte)
        reviewerByte, err = stub.GetState(PRKey)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        err = json.Unmarshal(reviewerByte, &reviewer)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }
        return errors.New(function + " : " + identifier + " has already registered the reviewer candidate, and the status is " + reviewer.Status +".")
    }

    reviewer = Reviewer { Key : PRKey, PaperKey : args[0], ReviewerID : identifier, Email : email, Status : "candidate" }
    reviewerByte, err =  json.Marshal(reviewer)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(PRKey, reviewerByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(RPKey, []byte(PRKey))
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

//args = [ paperKey, reviewerID, email ]
func (s *SmartContract) RegisterReviewer(stub shim.ChaincodeStubInterface, args []string) error {
    var PRKey string
    var PRKeyByte []byte
    var RPKey string
    var reviewerByte []byte
    var reviewer Reviewer
    var err error

    var paperKey = args[0];
    var reviewerID = args[1];
    var email = args[2];

    if len(args) != 3 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 3")
    }

    RPKey, err = ConvertReviewerKey(stub, paperKey, reviewerID)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    PRKeyByte, err = stub.GetState(RPKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    if PRKeyByte == nil {
        var reviewerCount int
        var identifier string

        identifier = args[1]
        reviewerCount, err = GetNumberOfReviewer(stub, args[0])
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        PRKey, err = CreateReviewerKeyWithPaperKey(stub, args[0], strconv.Itoa(reviewerCount + 1))
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        RPKey, err = ConvertReviewerKey(stub, PRKey, identifier)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        reviewer = Reviewer { Key : PRKey, PaperKey : args[0], ReviewerID : identifier, Email : email, Status : "selected" }
        reviewerByte, err =  json.Marshal(reviewer)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        err = stub.PutState(PRKey, reviewerByte)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        err = stub.PutState(RPKey, []byte(PRKey))
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }
    } else {
        PRKey = string(PRKeyByte)

        reviewerByte, err = stub.GetState(PRKey)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        err = json.Unmarshal(reviewerByte, &reviewer)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        if reviewer.Status != "candidate" {
            return errors.New(function + " : " + "The current reviewer's state is " + reviewer.Status + ", so the status can't be updated.")
        }

        reviewer.Status = "selected"

        if reviewer.Email == "" {
          reviewer.Email = email
        }

        reviewerByte, err = json.Marshal(reviewer)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        err = stub.PutState(PRKey, reviewerByte)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }
    }

    return nil
}

//args = [ paperKey ]
func (s *SmartContract) DeclineReview (stub shim.ChaincodeStubInterface, args []string) error {

    var err error
    var reviewerID string
    var reviewerByte []byte
    var reviewer Reviewer
    var RPKey string
    var PRKeyBytes []byte
    var PRKey string

    if len(args) != 1 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    reviewerID, err = GetIdentifier(stub)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    RPKey, err = ConvertReviewerKey(stub, args[0], reviewerID)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    PRKeyBytes, err = stub.GetState(RPKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    PRKey = string(PRKeyBytes)
    reviewerByte, err = stub.GetState(PRKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if reviewerByte == nil {
        return errors.New(function + " : " + reviewerID+ " is not reviewer of " + args[0] + " paper.")
    }

    err = json.Unmarshal(reviewerByte, &reviewer)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if reviewer.Status != "selected" {
        return errors.New(function + " : " + reviewerID + " is not selected as a reviewer.")
    }

    reviewer.Status = "declined"
    reviewerByte, err = json.Marshal(reviewer)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(PRKey, reviewerByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

//args = [ reviewerKey ]
func (s *SmartContract) RejectReviewer (stub shim.ChaincodeStubInterface, args []string) error {

    var paperKey string
    var reviewerByte []byte
    var reviewer Reviewer
    var err error

    if len(args) != 1 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    paperKey, err = CreatePaperKeyWithReviewerKey(stub, args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    reviewerByte, err = stub.GetState(args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if len(reviewerByte) == 0 {
        return errors.New(function + " : " + args[0] + " is not reviewer of " + paperKey + " paper.")
    }

    err = json.Unmarshal(reviewerByte, &reviewer)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    reviewer.Status = "rejected"
    reviewerByte, err = json.Marshal(reviewer)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(args[0], reviewerByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

//args = [ paperKey, reviewerNum ]
func (s *SmartContract) SelectReviewer(stub shim.ChaincodeStubInterface, args []string) error {

    var reviewerIter shim.StateQueryIteratorInterface
    var reviewerByte []byte
    var reviewer Reviewer
    var splitPaperKey []string
    var num int
    var count int
    var err error

    if len(args) != 2 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 2")
    }

    num, err = strconv.Atoi(args[1])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    _, splitPaperKey, err = stub.SplitCompositeKey(args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    reviewerIter, err = stub.GetStateByPartialCompositeKey("PR", splitPaperKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if(!reviewerIter.HasNext()) {
        return errors.New(function + " : " + "There is no reviewer of " + args[0] + " paper.")
    }

    count = 0
    for reviewerIter.HasNext() {
        if count >= num {
            break
        }

        reviewerKV, err := reviewerIter.Next()
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        err = json.Unmarshal(reviewerKV.Value, &reviewer)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        if reviewer.Status == "candidate" {
            reviewer.Status = "selected"
            reviewerByte, err = json.Marshal(reviewer)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            err = stub.PutState(reviewerKV.Key, reviewerByte)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            count++
        }
    }

    reviewerIter.Close()

    _, err = InvokeChaincode(stub, "paper", "UpdateStatus", []string { args[0], "reviewer_invited" })
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

//args = [ paperKey ]
func (s *SmartContract) StopRecruiting (stub shim.ChaincodeStubInterface, args []string) error {
    var err error

    if len(args) != 1 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    _, err = InvokeChaincode(stub, "paper", "UpdateStatus", []string { args[0], "reviewer_invited" })
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

//args = [ PRKey, status ]
func (s *SmartContract) UpdateStatus(stub shim.ChaincodeStubInterface, args []string) error {
    var paperKey string
    var reviewerByte []byte
    var reviewer Reviewer
    var err error

    if len(args) != 2 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 3")
    }

    paperKey, err = CreatePaperKeyWithReviewerKey(stub, args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    reviewerByte, err = stub.GetState(args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if len(reviewerByte) == 0 {
        return errors.New(function + " : " + args[0] + " is not reviewer of " + paperKey + " paper.")
    }

    err = json.Unmarshal(reviewerByte, &reviewer)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    reviewer.Status = args[1]

    reviewerByte, err = json.Marshal(reviewer)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = stub.PutState(args[0], reviewerByte)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    return nil
}

//args = [ reviewerID, email ]
func (s *SmartContract) UpdateEmail(stub shim.ChaincodeStubInterface, args []string) error {
    var PRKeyIter shim.StateQueryIteratorInterface
    var reviewerByte []byte
    var reviewer Reviewer
    var err error

    if len(args) != 2 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    PRKeyIter, err = stub.GetStateByPartialCompositeKey("RP", []string{ args[0] })
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }
    if !PRKeyIter.HasNext() {
        return nil
    }

    for PRKeyIter.HasNext() {
        PRKeyKV, err := PRKeyIter.Next()
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        reviewerByte, err = stub.GetState(string(PRKeyKV.Value))
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        err = json.Unmarshal(reviewerByte, &reviewer)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        reviewer.Email = args[1]
        reviewerByte, err = json.Marshal(reviewer)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        err = stub.PutState(reviewer.Key, reviewerByte)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }
    }

    PRKeyIter.Close()

    return nil
}

//args = [ paperKey ]
func (s *SmartContract) InitReviewer(stub shim.ChaincodeStubInterface, args []string) error {
    var splitPaperKey []string
    var reviewerIter shim.StateQueryIteratorInterface
    var reviewerByte []byte
    var reviewer Reviewer
    var err error

    if len(args) != 1 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    _, splitPaperKey, err = stub.SplitCompositeKey(args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    reviewerIter, err = stub.GetStateByPartialCompositeKey("PR", splitPaperKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    for reviewerIter.HasNext() {
        reviewerKV, err := reviewerIter.Next()
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        reviewerByte = reviewerKV.Value
        err = json.Unmarshal(reviewerByte, &reviewer)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        if reviewer.Status == "accept" || reviewer.Status == "submitted" {
            reviewer.Status = "selected"
            reviewerByte, err = json.Marshal(reviewer)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            err = stub.PutState(reviewerKV.Key, reviewerByte)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }
        }
    }
    reviewerIter.Close()

    return nil
}

//args = [ paperKey ]
func (s *SmartContract) DeleteReviewerWithPaperKey(stub shim.ChaincodeStubInterface, args []string) error {
    var splitPaperKey []string
    var reviewerIter shim.StateQueryIteratorInterface
    var err error

    if len(args) != 1 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    _, splitPaperKey, err = stub.SplitCompositeKey(args[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    reviewerIter, err = stub.GetStateByPartialCompositeKey("PR", splitPaperKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    for reviewerIter.HasNext() {
        reviewerKV, err := reviewerIter.Next()
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }

        err = stub.DelState(reviewerKV.Key)
        if err != nil {
            return errors.New(function + " : " + err.Error())
        }
    }
    reviewerIter.Close()

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
