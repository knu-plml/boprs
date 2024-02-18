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
//    var err error
//
//    userType = REVIEWER
//
//    switch function {
//        //author of paper
//        // args[0] = paperKey
//        case "AuthorRateReviewer" :
//            paperKey = args[0]
//            paper, err = GetPaper(stub, paperKey)
//            if err != nil {
//                return userType, errors.New("AccessControl"+ " : " + err.Error())
//            }
//
//            if paper.Status != "rejected" && paper.Status != "accepted" {
//                return userType, errors.New("AccessControl" + " : " + " The review of " + args[0] + " paper is not over.")
//            }
//
//            err = CertifyAuthorWithPaperKey(stub, paperKey)
//            if err == nil {
//                userType = AUTHOR
//            } else {
//                return userType, errors.New("AccessControl" + " : " + err.Error())
//            }
//
//            return userType, nil
//
//        //editor of paper
//        // args[0] = paperKey
//        case "EOPRateReviewer" :
//            paperKey = args[0]
//            paper, err = GetPaper(stub, paperKey)
//            if err != nil {
//                return userType, errors.New("AccessControl"+ " : " + err.Error())
//            }
//
//            if paper.Status != "rejected" && paper.Status != "accepted" {
//                return userType, errors.New("AccessControl" + " : " + " The review of " + args[0] + " paper is not over.")
//            }
//
//            err = CertifyEOPWithPaperKey(stub, paperKey)
//            if err == nil {
//                userType = EOP
//            } else {
//                return userType, errors.New("AccessControl" + " : " + err.Error())
//            }
//
//            return userType, nil
//
//        //Anyone
//        case "QueryRating" :
//            paperKey, err = CreatePaperKeyWithRatingKey(stub, args[0])
//            if err != nil {
//                return userType, errors.New("AccessControl" + " : " + err.Error())
//            }
//
//            userType, err = GetUserTypeWithPaperKey(stub, paperKey)
//            if err != nil {
//                return userType, errors.New("AccessControl" + " : " + err.Error())
//            }
//
//            return userType, nil
//
//        case "QueryRatingWithPaperKey" :
//            paperKey = args[0]
//            userType, err = GetUserTypeWithPaperKey(stub, paperKey)
//            if err != nil {
//                return userType, errors.New("AccessControl" + " : " + err.Error())
//            }
//
//            return userType, nil
//
//        case "QueryRatingWithReviewerID" :
//            reviewerID := args[0]
//            err = CertifyIdentifier(stub, reviewerID)
//            if err == nil {
//                userType = SELF
//            } else {
//                err = CertifyEditor(stub)
//                if err == nil {
//                    userType = EDITOR
//                }
//            }
//
//            return userType, nil
//
//        case "QuerySimpleRatingWithPaperKey" :
//            paperKey = args[0]
//            paper, err = GetPaper(stub, paperKey)
//            if err != nil {
//                return userType, errors.New("AccessControl"+ " : " + err.Error())
//            }
//
//            if paper.Status != "rejected" && paper.Status != "accepted" {
//                return userType, errors.New("AccessControl" + " : " + " The review of " + args[0] + " paper is not over.")
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
//                    return userType, errors.New("AccessControl" + " : " + err.Error())
//                }
//            }
//
//            return userType, nil
//
//        default :
//            return userType, errors.New("AccessControl"+ " : " + function + " is Invalid Smart Contract function name.")
//        }
//}
//
//func ApplyAcceptanceModel(stub shim.ChaincodeStubInterface, function string, userType UserType, target []byte, args []string) ([]byte, error) {
//    var paperKey string
//    var organization string
//    var identifier string
//    var acceptanceModel *AcceptanceModel
//    var rating Rating
//    var ratingArray []Rating
//    var ratingByte []byte
//    var buff bytes.Buffer
//    var result []byte
//    var writeOne bool
//    var err error
//
//
//    result = target
//    writeOne = false
//
//
//    if function == "QuerySimpleRatingWithPaperKey" && (userType == EOP || userType == AUTHOR){
//        return target, nil
//    }
//
//    switch function {
//        // return Rating
//        case "QueryRating" :
//            paperKey, err = CreatePaperKeyWithRatingKey(stub, args[0])
//            if err != nil {
//                return target, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
//            }
//
//            organization, err = GetOrganizationFromPaperKey(stub, paperKey)
//            if err != nil {
//                return target, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
//            }
//
//            acceptanceModel, err = GetAcceptanceModel(stub, organization)
//            if err != nil {
//                return target, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
//            }
//
//            if userType != AUTHOR && userType != EOP && userType != EOO && acceptanceModel.OI == "0" {
//                err = json.Unmarshal(target, &rating)
//                if err != nil {
//                    return target, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
//                }
//
//                rating.ReviewerID = ""
//                result, err = json.Marshal(rating)
//                if err != nil {
//                    return target, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
//                }
//            }
//
//            return result, nil
//
//        // return Rating Array
//        case "QueryRatingWithPaperKey" :
//            paperKey = args[0]
//            organization, err = GetOrganizationFromPaperKey(stub, paperKey)
//            if err != nil {
//                return target, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
//            }
//
//            acceptanceModel, err = GetAcceptanceModel(stub, organization)
//            if err != nil {
//                return target, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
//            }
//
//            if userType != AUTHOR && userType != EOP && userType != EOO && acceptanceModel.OI == "0" {
//                err = json.Unmarshal(target, &ratingArray)
//                if err != nil {
//                    return target, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
//                }
//                if len(ratingArray) == 0 {
//                    return target, nil
//                }
//
//                identifier, err = GetIdentifier(stub)
//                if err != nil {
//                    return target, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
//                }
//
//                buff.WriteString("[")
//                for _, rating := range ratingArray {
//                    if identifier != rating.ReviewerID {
//                        rating.ReviewerID = ""
//                    }
//
//                    ratingByte, err = json.Marshal(rating)
//                    if err != nil {
//                        return target, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
//                    }
//
//                    if writeOne {
//                        buff.WriteString(",")
//                    }
//
//                    writeOne = true
//                    buff.Write(ratingByte)
//                }
//                buff.WriteString("]")
//                return buff.Bytes(), nil
//            } else {
//                return target, nil
//            }
//
//        case "QueryRatingWithReviewerID" :
//            if userType == SELF {
//                return target, nil
//            }
//
//            err = json.Unmarshal(target, &ratingArray)
//            if err != nil {
//                return target, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
//            }
//            if len(ratingArray) == 0 {
//                return target, nil
//            }
//
//            buff.WriteString("[")
//            for _, rating := range ratingArray {
//                paperKey = rating.PaperKey
//                organization, err = GetOrganizationFromPaperKey(stub, paperKey)
//                if err != nil {
//                    return target, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
//                }
//
//                acceptanceModel, err = GetAcceptanceModel(stub, organization)
//                if err != nil {
//                    return target, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
//                }
//
//                userType, err = GetUserTypeWithPaperKey(stub, paperKey)
//                if err != nil {
//                    return target, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
//                }
//
//                if userType == AUTHOR || userType == EOP || userType == EOO || acceptanceModel.OI != "0" {
//                    ratingByte, err = json.Marshal(rating)
//                    if err != nil {
//                        return target, errors.New("ApplyAcceptanceModel" + " : " + err.Error())
//                    }
//
//                    if writeOne {
//                        buff.WriteString(",")
//                    }
//
//                    writeOne = true
//                    buff.Write(ratingByte)
//                }
//            }
//            buff.WriteString("]")
//            return buff.Bytes(), nil
//
//        default :
//            return target, errors.New("ApplyAcceptanceModel"+ " : " + function + " is Invalid Smart Contract function name.")
//
//    }
//}

func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
    var args []string
//    var userType UserType
    var result []byte = nil
    var maskedResult []byte = nil
    var err error

    //get function name and argument list
    function, args = stub.GetFunctionAndParameters()

    //access control
//    userType, err = AccessControl(stub, function, args)
//    if err != nil {
//        return shim.Error(err.Error())
//    }

    switch function {
        // No Return
        case "AuthorRateReviewer" :
            err = s.AuthorRateReviewer(stub, args)
        case "EOPRateReviewer" :
            err = s.EOPRateReviewer(stub, args)

        // Rating
        case "QueryRating" :
            result, err = s.QueryRating(stub, args)

        // Rating Array
        case "QueryRatingWithPaperKey" :
            result, err = s.QueryRatingWithPaperKey(stub, args)
        case "QueryRatingWithReviewerID" :
            result, err = s.QueryRatingWithReviewerID(stub, args)
        case "QuerySimpleRatingWithPaperKey" :
            result, err = s.QuerySimpleRatingWithPaperKey(stub, args)
    }

    if err != nil {
        return shim.Error(err.Error())
    }

    maskedResult = result
//    if result != nil {
//        maskedResult, err = ApplyAcceptanceModel(stub, function, userType, result, args)
//        if err != nil {
//            return shim.Error(err.Error())
//        }
//    }

    return shim.Success(maskedResult)
}

//args = [ PaperKey, ReviewerIndex Array, Rating Array ]
func (s *SmartContract) AuthorRateReviewer(stub shim.ChaincodeStubInterface, args []string) error {
    var indexArray []string
    var ratingArray []string
    var PRKey string
    var RatingKey string
    var rating Rating
    var ratingByte []byte
    var err error

    if len(args) != 3 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 3")
    }

    err = json.Unmarshal([]byte(args[1]), &indexArray)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = json.Unmarshal([]byte(args[2]), &ratingArray)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    if len(indexArray) != len(ratingArray) {
        return errors.New(function + " : " + "ReviewerID Array and Rating Array have differnt length.")
    }

    PRKey, err = CreateReviewerKeyWithPaperKey(stub, args[0], indexArray[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    RatingKey, err = CreateRatingKeyWithPaperKey(stub, args[0], PRKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    ratingByte, err = stub.GetState(RatingKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    if len(ratingByte) == 0 {
        for i, index := range indexArray {
            PRKey, err = CreateReviewerKeyWithPaperKey(stub, args[0], index)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            RatingKey, err = CreateRatingKeyWithPaperKey(stub, args[0], PRKey)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            rating = Rating { Key : RatingKey, ReviewerID : "", PRKey : PRKey, PaperKey : args[0], AuthorRating : ratingArray[i], EditorRating : "" }

            ratingByte, err = json.Marshal(rating)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            err = stub.PutState(RatingKey, ratingByte)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }
        }
    } else {
        for i, index := range indexArray {
            PRKey, err = CreateReviewerKeyWithPaperKey(stub, args[0], index)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            RatingKey, err = CreateRatingKeyWithPaperKey(stub, args[0], PRKey)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            ratingByte, err = stub.GetState(RatingKey)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            err = json.Unmarshal(ratingByte, &rating)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            rating.AuthorRating = ratingArray[i]
            ratingByte, err = json.Marshal(rating)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            err = stub.PutState(RatingKey, ratingByte)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }
        }
    }

    return nil
}

//args = [ PaperKey, ReviewerID Array, Rating Array ]
func (s *SmartContract) EOPRateReviewer(stub shim.ChaincodeStubInterface, args []string) error {
    var reviewerIDArray []string
    var ratingArray []string
    var index string
    var RPKey string
    var PRKey string
    var RatingKey string
    var RRatingKey string
    var rating Rating
    var ratingByte []byte
    var err error

    if len(args) != 3 {
        return errors.New(function + " : " + "Incorrect number of arguments. Expecting 3")
    }

    err = json.Unmarshal([]byte(args[1]), &reviewerIDArray)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    err = json.Unmarshal([]byte(args[2]), &ratingArray)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    if len(reviewerIDArray) != len(ratingArray) {
        return errors.New(function + " : " + "ReviewerID Array and Rating Array have differnt length.")
    }

    RPKey, err = ConvertReviewerKey(stub, args[0], reviewerIDArray[0])
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    index, err = GetReviewerIndex(stub, RPKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    PRKey, err = CreateReviewerKeyWithPaperKey(stub, args[0], index)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    RatingKey, err = CreateRatingKeyWithPaperKey(stub, args[0], PRKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    ratingByte, err = stub.GetState(RatingKey)
    if err != nil {
        return errors.New(function + " : " + err.Error())
    }

    if len(ratingByte) == 0 {
        for i, reviewerID := range reviewerIDArray {
            RPKey, err = ConvertReviewerKey(stub, args[0], reviewerID)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            index, err = GetReviewerIndex(stub, RPKey)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            PRKey, err = CreateReviewerKeyWithPaperKey(stub, args[0], index)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            RatingKey, err = CreateRatingKeyWithPaperKey(stub, args[0], PRKey)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            RRatingKey, err = ConvertRatingKey(stub, RatingKey, reviewerID)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            rating = Rating { Key : RatingKey, ReviewerID : reviewerID, PRKey : PRKey, PaperKey : args[0], AuthorRating : "", EditorRating : ratingArray[i] }

            ratingByte, err = json.Marshal(rating)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            err = stub.PutState(RatingKey, ratingByte)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            err = stub.PutState(RRatingKey, []byte(RatingKey))
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }
        }
    } else {
        for i, reviewerID := range reviewerIDArray {
            RPKey, err = ConvertReviewerKey(stub, args[0], reviewerID)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            index, err = GetReviewerIndex(stub, RPKey)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            PRKey, err = CreateReviewerKeyWithPaperKey(stub, args[0], index)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            RatingKey, err = CreateRatingKeyWithPaperKey(stub, args[0], PRKey)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            ratingByte, err = stub.GetState(RatingKey)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            err = json.Unmarshal(ratingByte, &rating)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            rating.EditorRating = ratingArray[i]
            rating.ReviewerID = reviewerID
            ratingByte, err = json.Marshal(rating)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }

            err = stub.PutState(RatingKey, ratingByte)
            if err != nil {
                return errors.New(function + " : " + err.Error())
            }
        }
    }

    return nil
}

//args = [ RatingKey ]
func (s *SmartContract) QueryRating(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var ratingByte []byte
    var err error

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    ratingByte, err = stub.GetState(args[0])
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if len(ratingByte) == 0 {
        return nil, errors.New(function + " : " + args[0] + " rating is not exist.")
    }

    return ratingByte, nil
}

//args = [ paperKey ]
func (s *SmartContract) QueryRatingWithPaperKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var ratingIter shim.StateQueryIteratorInterface
    var splitPaperKey []string
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

    ratingIter, err = stub.GetStateByPartialCompositeKey("PR", splitPaperKey)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !ratingIter.HasNext() {
        return []byte("[]"), nil
    }

    flag = false
    buff.WriteString("[")
    for ratingIter.HasNext() {
        ratingKV, err := ratingIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if flag {
            buff.WriteString(",")
        } else {
            flag = true
        }

        buff.Write(ratingKV.Value)
    }

    buff.WriteString("]")
    ratingIter.Close()

    return buff.Bytes(), nil
}

//args = [ reviewerID ]
func (s *SmartContract) QueryRatingWithReviewerID(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var ratingIter shim.StateQueryIteratorInterface
    var ratingByte []byte
    var buff bytes.Buffer
    var flag bool
    var err error

    if len(args) != 1 {
        return nil, errors.New(function + " : " + "Incorrect number of arguments. Expecting 1")
    }

    ratingIter, err = stub.GetStateByPartialCompositeKey("RP", []string{ args[0] })
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !ratingIter.HasNext() {
        return []byte("[]"), nil
    }

    flag = false
    buff.WriteString("[")
    for ratingIter.HasNext() {
        ratingKV, err := ratingIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if flag {
            buff.WriteString(",")
        } else {
            flag = true
        }

        ratingByte, err = stub.GetState(string(ratingKV.Value))
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        buff.Write(ratingByte)
    }

    buff.WriteString("]")
    ratingIter.Close()

    return buff.Bytes(), nil
}

//args = [ paperKey ]
func (s *SmartContract) QuerySimpleRatingWithPaperKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var ratingIter shim.StateQueryIteratorInterface
    var rating Rating
    var splitPaperKey []string
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

    ratingIter, err = stub.GetStateByPartialCompositeKey("PR", splitPaperKey)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if !ratingIter.HasNext() {
        return []byte("[]"), nil
    }

    flag = false
    buff.WriteString("[")
    for ratingIter.HasNext() {
        ratingKV, err := ratingIter.Next()
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        err = json.Unmarshal(ratingKV.Value, &rating)
        if err != nil {
            return nil, errors.New(function + " : " + err.Error())
        }

        if flag {
            buff.WriteString(",")
        } else {
            flag = true
        }

        buff.WriteString("{\"ORCID\" : \"")
        buff.Write([]byte(rating.ReviewerID))
        buff.WriteString("\",\"AuthorRating\" : \"")
        buff.Write([]byte(rating.AuthorRating))
        buff.WriteString("\",\"EditorRating\" : \"")
        buff.Write([]byte(rating.EditorRating))
        buff.WriteString("\"}")
    }

    buff.WriteString("]")
    ratingIter.Close()

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
