package main
import(
//    "fmt"
    "strconv"
    "errors"
    "encoding/json"
    "strings"

    "github.com/hyperledger/fabric-chaincode-go/shim"
    sc "github.com/hyperledger/fabric-protos-go/peer"
    "github.com/hyperledger/fabric-chaincode-go/pkg/cid"
)

func OICheck(paper *Paper, acceptanceModel *AcceptanceModel) bool {
    var open bool

    open = true

    switch acceptanceModel.OI {
    case "0" :
        open = false

    case "1" :
        if paper.Status != "under_review" && paper.Status != "under_decision" && paper.Status != "revise" && paper.Status != "accepted" && paper.Status != "rejected" {
            open = false
        }

    case "2" :
        if paper.Status != "accepted" && paper.Status != "rejected" {
            open = false
        }

    case "3" :
        if paper.Status != "accepted" {
            open = false
        }
    }

    return open
}

func ORCheck(paper *Paper, acceptanceModel *AcceptanceModel) bool {
    var open bool

    open = true

    switch acceptanceModel.OR {
    case "0" :
        open = false

    case "4", "5" :
        if paper.Status != "under_review" && paper.Status != "under_decision" && paper.Status != "revise" && paper.Status != "accepted" && paper.Status != "rejected" {
            open = false
        }

    case "6", "7" :
        if paper.Status != "accepted" && paper.Status != "rejected" {
            open = false
        }
    }

    return open
}

func OCCheck(paper *Paper, acceptanceModel *AcceptanceModel) bool {
    var open bool

    open = true

    switch acceptanceModel.OC {
    case "0" :
        open = false

    case "1" :
        if paper.Status != "accepted" && paper.Status != "rejected" {
            open = false
        }
    }

    return open
}

func ONCheck(paper *Paper, acceptanceModel *AcceptanceModel) bool {
    var open bool

    open = true

    switch acceptanceModel.ON {
    case "0" :
        open = false

    case "4" :
        if paper.Status != "under_review" && paper.Status != "under_decision" && paper.Status != "revise" && paper.Status != "accepted" && paper.Status != "rejected" {
            open = false
        }

    case "7" :
        if paper.Status != "accepted" && paper.Status != "rejected" {
            open = false
        }
    }

    return open
}

func InvokeChaincode(stub shim.ChaincodeStubInterface, chainName string, funcName string, args []string) ([]byte, error) {
    const function string = "InvokeChaincode"
    var ChannelName string = stub.GetChannelID()
    var functionByte []byte
    var arg string
    var argsByte [][]byte
    var response sc.Response

    functionByte = []byte (funcName)

    argsByte = [][]byte { functionByte }

    for _, arg = range args {
        argsByte = append(argsByte, []byte (arg))
    }

    response = stub.InvokeChaincode(chainName, argsByte, ChannelName)
    if response.Status != shim.OK {
//        return nil, fmt.Errorf(function + "(" + "ChannelName:" + ChannelName + ", chainName:" + chainName + ", funcName:" + funcName + ", args:[" + strings.Join(args, ", ") + "]) : " + response.GetMessage())
        return nil, errors.New(function + "(" + "ChannelName:" + ChannelName + ", chainName:" + chainName + ", funcName:" + funcName + ", args:[" + strings.Join(args, ", ") + "]) : " + response.GetMessage())
    }
    return response.GetPayload(), nil
}

func GetIdentifier (stub shim.ChaincodeStubInterface) (string, error) {
    const function string = "GetIdentifier"

    orcid, ok, err := cid.GetAttributeValue(stub, "identifier")
    if err != nil {
        return "", err
    }
    if !ok {
        return "", errors.New(function + " : " + "The client identity does not possess the attribute")
    }

    return orcid, nil
}

func GetEmail (stub shim.ChaincodeStubInterface) (string, error) {
    const function string = "GetEmail"

    email, ok, err := cid.GetAttributeValue(stub, "email")
    if err != nil {
        return "", err
    }
    if !ok {
        return "", errors.New(function + " : " + "The client email does not possess the attribute")
    }

    return email, nil
}

func GetMSPID(stub shim.ChaincodeStubInterface) (string, error) {
    const function string = "GetMSPID"

    mspKey, err := cid.GetMSPID(stub)
    if err != nil {
        return "", err
    }

    return mspKey, nil
}

func GetPaperByte (stub shim.ChaincodeStubInterface, paperKey string) ([]byte, error) {
    const function string = "GetPaperByte"
    var paperByte []byte
    var err error

    paperByte, err = InvokeChaincode(stub, "paper", "QueryPaper", []string { paperKey })
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }
    if len(paperByte) == 0 {
        return nil, errors.New(function + " : " + "Can't cound paper " + paperKey + ".")
    }

    return paperByte, nil
}

func GetPaper (stub shim.ChaincodeStubInterface, paperKey string) (*Paper, error) {
    const function string = "GetPaper"
    var paperByte []byte
    var paper Paper
    var err error

    paperByte, err = GetPaperByte(stub, paperKey)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }

    err = json.Unmarshal(paperByte, &paper)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }

    return &paper, nil
}

func GetAcceptanceModel (stub shim.ChaincodeStubInterface, organization string) (*AcceptanceModel, error) {
    const function string = "GetAcceptanceModel"
    var acceptanceModelByte []byte
    var acceptanceModel AcceptanceModel
    var err error

    acceptanceModelByte, err = InvokeChaincode(stub, "acceptanceModel", "QueryAcceptanceModel", []string { organization })
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }

    err = json.Unmarshal(acceptanceModelByte, &acceptanceModel)
    if err != nil {
        return nil, errors.New(function + " : " + err.Error())
    }

    return &acceptanceModel, nil
}

func CountReviewerStatus (stub shim.ChaincodeStubInterface, paperKey string, status string) (int, error) {
    const function string = "CountReviewerStatus"
    var reviewersByte []byte
    var reviewersArray []Reviewer
    var count int
    var err error

//    err = CertifyEOPWithPaperKey(stub, paperKey)
//    if err != nil {
//        return -1, errors.New(function + " : " + err.Error())
//    }

    reviewersByte, err = InvokeChaincode(stub, "reviewer", "QueryReviewerWithPaperKey", []string { paperKey })
    if err != nil {
        return -1, errors.New(function + " : " + err.Error())
    }

    err = json.Unmarshal(reviewersByte, &reviewersArray)
    if err != nil {
        return -1, errors.New(function + " : " + err.Error())
    }

    count = 0
    for _, reviewer := range reviewersArray {
        if reviewer.Status == status {
            count++
        }
    }

    return count, nil
}

func CreatePaperKey(stub shim.ChaincodeStubInterface, organization string, paperID string) (string, error) {
    const function string = "CreatePaperKey"

    paperKey, err := stub.CreateCompositeKey("OI", []string { organization, paperID })
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return paperKey, nil
}

func CreatePaperFileKeyWithPaperKey(stub shim.ChaincodeStubInterface, paperKey string, revision int) (string, error) {
    _, splitPaperKey, err := stub.SplitCompositeKey(paperKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    paperFileKey, err := stub.CreateCompositeKey("PF", []string { splitPaperKey[0], splitPaperKey[1], strconv.Itoa(revision) })
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return paperFileKey, nil
}

func CreatePaperKeyWithPaperFileKey(stub shim.ChaincodeStubInterface, paperFileKey string) (string, error) {
    const function string = "CreatePaperKeyWithPaperFileKey"

    _, splitPaperFileKey, err := stub.SplitCompositeKey(paperFileKey)

    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    paperKey, err := stub.CreateCompositeKey("OI", []string { splitPaperFileKey[0], splitPaperFileKey[1] })
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return paperKey, nil
}

func CreatePaperKeyWithContractKey(stub shim.ChaincodeStubInterface, contractKey string) (string, error) {
    const function string = "CreatePaperKeyWithContractKey"

    _, splitContractKey, err := stub.SplitCompositeKey(contractKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    paperKey, err := stub.CreateCompositeKey("OI", []string { splitContractKey[0], splitContractKey[1] })
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return paperKey, nil
}

func CreatePaperKeyWithReviewerKey(stub shim.ChaincodeStubInterface, reviewerKey string) (string, error) {
    const function string = "CreatePaperKeyWithReviewerKey"

    _, splitedReviewerKey, err := stub.SplitCompositeKey(reviewerKey)

    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    paperKey, err := stub.CreateCompositeKey("OI", []string { splitedReviewerKey[0], splitedReviewerKey[1] })
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return paperKey, nil
}

func CreatePaperKeyWithSignatureKey(stub shim.ChaincodeStubInterface, signatureKey string) (string, error) {
    const function string = "CreatePaperKeyWithSignatureKey"

    _, splitSignatureKey, err := stub.SplitCompositeKey(signatureKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    paperKey, err := stub.CreateCompositeKey("OI", []string { splitSignatureKey[0], splitSignatureKey[1] })
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return paperKey, nil
}

func CreatePaperKeyWithCommentKey(stub shim.ChaincodeStubInterface, commentKey string) (string, error) {
    const function string = "CreatePaperKeyWithCommentKey"

    _, splitCommentKey, err := stub.SplitCompositeKey(commentKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    paperKey, err := stub.CreateCompositeKey("OI", []string { splitCommentKey[0], splitCommentKey[1] })
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return paperKey, nil
}

func CreatePaperKeyWithRatingKey(stub shim.ChaincodeStubInterface, ratingKey string) (string, error) {
    const function string = "CreatePaperKeyWithRatingKey"

    _, splitRatingKey, err := stub.SplitCompositeKey(ratingKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    paperKey, err := stub.CreateCompositeKey("OI", []string { splitRatingKey[0], splitRatingKey[1] })
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return paperKey, nil
}

func CreatePaperKeyWithMessageKey(stub shim.ChaincodeStubInterface, messageKey string) (string, error) {
    return CreatePaperKeyWithContractKey(stub, messageKey)
}

func CreateContractKeyWithPaperKey(stub shim.ChaincodeStubInterface, paperKey string, round string) (string, error) {
    const function string = "CreateContractKeyWithPaperKey"

    _, splitPaperKey, err := stub.SplitCompositeKey(paperKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    contractKey, err := stub.CreateCompositeKey("", []string { splitPaperKey[0], splitPaperKey[1], round })
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return contractKey, nil
}

func CreateContractKeyWithCommentKey(stub shim.ChaincodeStubInterface, commentKey string) (string, error) {
    const function string = "CreateContractKeyWithCommentKey"

    _, splitCommentKey, err := stub.SplitCompositeKey(commentKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    contractKey, err := stub.CreateCompositeKey("", []string { splitCommentKey[0], splitCommentKey[1], splitCommentKey[2] })
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return contractKey, nil
}

func CreateReviewerKeyWithPaperKey(stub shim.ChaincodeStubInterface, paperKey string, reviewerIndex string) (string, error) {
    const function string = "CreateReviewerKeyWithPaperKey"

    _, splitPaperKey, err := stub.SplitCompositeKey(paperKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    reviewerKey, err := stub.CreateCompositeKey("PR", []string { splitPaperKey[0], splitPaperKey[1], reviewerIndex })
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return reviewerKey, nil
}

func CreateSignatureKeyWithContractKey(stub shim.ChaincodeStubInterface, contractKey string, reviewerID string) (string, error) {
    const function string = "CreateSignatureKeyWithContractKey"

    _, splitContractKey, err := stub.SplitCompositeKey(contractKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    signatureKey, err := stub.CreateCompositeKey("SIG", []string { splitContractKey[0], splitContractKey[1], splitContractKey[2], reviewerID })
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return signatureKey, nil
}

func CreateCommentKeyWithContractKey(stub shim.ChaincodeStubInterface, contractKey string, reviewerIndex string) (string, error) {
    const function string = "CreateCommentKeyWithContractKey"

    _, splitContractKey, err := stub.SplitCompositeKey(contractKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    commentKey, err := stub.CreateCompositeKey("CR", []string { splitContractKey[0], splitContractKey[1], splitContractKey[2], reviewerIndex })
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return commentKey, nil
}

func CreateCommentKeyWithRevisionNoteKey(stub shim.ChaincodeStubInterface, revisionNoteKey string) (string, error) {
    const function string = "CreateCommentKeyWithRevisionNoteKey"

    _, splitedRevisionNoteKey, err := stub.SplitCompositeKey(revisionNoteKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    commentKey, err := stub.CreateCompositeKey("CR", splitedRevisionNoteKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return commentKey, nil
}

func CreateRevisionNoteKeyWithCommentKey(stub shim.ChaincodeStubInterface, commentKey string) (string, error) {
    const function string = "CreateRevisionNoteKeyWithCommentKey"

    _, splitCommentKey, err := stub.SplitCompositeKey(commentKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    revisionNoteKey, err := stub.CreateCompositeKey("RN", splitCommentKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return revisionNoteKey, nil
}

func CreateRatingKeyWithPaperKey(stub shim.ChaincodeStubInterface, paperKey string, PRKey string) (string, error) {
    const function string = "CreateRatingKeyWithPaperKey"

    reviewerIndex, err := GetReviewerIndex(stub, PRKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    _, splitPaperKey, err := stub.SplitCompositeKey(paperKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    ratingKey, err := stub.CreateCompositeKey("PR", []string { splitPaperKey[0], splitPaperKey[1], reviewerIndex })
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return ratingKey, nil
}

func CreateMessageKey(stub shim.ChaincodeStubInterface, key string, index string) (string, error) {
    const function string = "CreateMessageKey"

    _, splitedKey, err := stub.SplitCompositeKey(key)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    splitedKey = append(splitedKey, index)
    messageKey, err := stub.CreateCompositeKey("MR", splitedKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return messageKey, nil
}

func CreateMessageIndexKey(stub shim.ChaincodeStubInterface, key string) (string, error) {
    const function string = "CreateMessageIndexKey"

    _, splitedKey, err := stub.SplitCompositeKey(key)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    messageKey, err := stub.CreateCompositeKey("MI", splitedKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return messageKey, nil
}

func ConvertPaperKey(stub shim.ChaincodeStubInterface, paperKey string, authorID string) (string, error) {
    const function string = "ConvertPaperKey"

    docType, splitPaperKey, err := stub.SplitCompositeKey(paperKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    if docType == "OI" {
        newPaperKey, err := stub.CreateCompositeKey("AOI", []string { authorID, splitPaperKey[0], splitPaperKey[1] })
        if err != nil {
            return "", errors.New(function + " : " + err.Error())
        }

        return newPaperKey, nil
    } else {
        return "", errors.New(function + " : " + paperKey + " is not OI PaperKey.")
    }
}

func ConvertReviewerKey(stub shim.ChaincodeStubInterface, Key string, reviewerID string) (string, error) {
    const function string = "ConvertReviewerKey"

    docType, splitKey, err := stub.SplitCompositeKey(Key)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    if docType == "PR" || docType == "OI" {
        newReviewerKey, err := stub.CreateCompositeKey("RP", []string { reviewerID, splitKey[0], splitKey[1] })
        if err != nil {
            return "", errors.New(function + " : " + err.Error())
        }

        return newReviewerKey, nil
    } else {
        return "", errors.New(function + " : " + Key + " can't Convert to RP ReviewerID.")
    }
}

func ConvertCommentKey(stub shim.ChaincodeStubInterface, commentKey string, reviewerID string) (string, error) {
    const function string = "ConvertCommentKey"

    docType, splitCommentKey, err := stub.SplitCompositeKey(commentKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    if docType == "CR" {
        newCommentKey, err := stub.CreateCompositeKey("RC", []string { reviewerID, splitCommentKey[0], splitCommentKey[1], splitCommentKey[2], splitCommentKey[3] })
        if err != nil {
            return "", errors.New(function + " : " + err.Error())
        }

        return newCommentKey, nil
    } else {
        return "", errors.New(function + " : " + commentKey + " is not CommentKey.")
    }
}

func ConvertRatingKey(stub shim.ChaincodeStubInterface, Key string, reviewerID string) (string, error) {
    const function string = "ConvertRatingKey"

    docType, splitKey, err := stub.SplitCompositeKey(Key)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    if docType == "PR" || docType == "OI" {
        newRatingKey, err := stub.CreateCompositeKey("RP", []string { reviewerID, splitKey[0], splitKey[1] })
        if err != nil {
            return "", errors.New(function + " : " + err.Error())
        }

        return newRatingKey, nil
    } else {
        return "", errors.New(function + " : " + Key + " can't Convert to RP RatingKey.")
    }
}

func ConvertRevisionNoteKey(stub shim.ChaincodeStubInterface, revisionNoteKey string, reviewerID string) (string, error) {
    const function string = "ConvertRevisionNoteKey"

    docType, splitRevisionNoteKey, err := stub.SplitCompositeKey(revisionNoteKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    if docType == "RN" {
        newRevisionNoteKey, err := stub.CreateCompositeKey("RR", []string { reviewerID, splitRevisionNoteKey[0], splitRevisionNoteKey[1], splitRevisionNoteKey[2], splitRevisionNoteKey[3] })
        if err != nil {
            return "", errors.New(function + " : " + err.Error())
        }

        return newRevisionNoteKey, nil
    } else {
        return "", errors.New(function + " : " + revisionNoteKey + " is not RevisionNoteKey.")
    }
}

func ConvertMessageKey(stub shim.ChaincodeStubInterface, key string, reviewerID string) (string, error) {
    const function string = "ConvertMessageKey"

    docType, splitedKey, err := stub.SplitCompositeKey(key)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    if docType == "MR" {
        splitedKey = append([]string{ reviewerID }, splitedKey...)
        newMessageKey, err := stub.CreateCompositeKey("RM", splitedKey)
        if err != nil {
            return "", errors.New(function + " : " + err.Error())
        }

        return newMessageKey, nil
    } else {
        return "", errors.New(function + " : " + key + " is not MessageKey.")
    }
}

func GetReviewerIndex(stub shim.ChaincodeStubInterface, reviewerKey string) (string, error) {
    const function string = "GetReviewerIndex"
    var splitedReviewerKey []string
    var docType string
    var PRKeyBytes []byte
    var PRKey string
    var err error

    docType, splitedReviewerKey, err = stub.SplitCompositeKey(reviewerKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    if docType == "PR" {
        return splitedReviewerKey[2], nil
    } else if docType == "RP" {
        PRKeyBytes, err = InvokeChaincode(stub, "reviewer", "GetPRKey", []string { reviewerKey })
        if err != nil {
            return "", errors.New(function + " : " + err.Error())
        }

        PRKey = string(PRKeyBytes)
        _, splitedReviewerKey, err = stub.SplitCompositeKey(PRKey)
        if err != nil {
            return "", errors.New(function + " : " + err.Error())
        }

        return splitedReviewerKey[2], nil
    } else if docType == "CR" {
        return splitedReviewerKey[3], nil
    } else {
        return "", errors.New(function + " : " + reviewerKey + " can't convert to ReviewerID.")
    }
}

func GetOrganizationFromPaperKey (stub shim.ChaincodeStubInterface, paperKey string) (string, error) {
    var function string = "GetOrganizationFromPaperKey"
    var docType string
    var splitedPaperKey []string
    var err error

    docType, splitedPaperKey, err = stub.SplitCompositeKey(paperKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    if docType == "OI" {
        return splitedPaperKey[0], nil
    } else if docType == "AOI" {
        return splitedPaperKey[1], nil
    } else {
        return "", errors.New(function + " : " + "Organization cannot be extracted from this key.")
    }
}

func GetRoundFromCommentKey (stub shim.ChaincodeStubInterface, commentKey string) (string, error) {
    var splitedCommentKey []string
    var err error

    _, splitedCommentKey, err = stub.SplitCompositeKey(commentKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return splitedCommentKey[2], nil
}

func GetRoundFromContractKey (stub shim.ChaincodeStubInterface, contractKey string) (string, error) {
    var splitedContractKey []string
    var err error

    _, splitedContractKey, err = stub.SplitCompositeKey(contractKey)
    if err != nil {
        return "", errors.New(function + " : " + err.Error())
    }

    return splitedContractKey[2], nil
}
