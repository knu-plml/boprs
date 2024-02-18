package main
import (
    "errors"
    "encoding/json"

    "github.com/hyperledger/fabric-chaincode-go/shim"
    "github.com/hyperledger/fabric-chaincode-go/pkg/cid"
)

func GetUserTypeWithPaperKey (stub shim.ChaincodeStubInterface, paperKey string) (UserType, error) {
    var paper *Paper
    var err error

    if paperKey == "" {
        return REVIEWER, nil
    }

    paper, err = GetPaper(stub, paperKey)
    if err != nil {
        return REVIEWER, errors.New("GetUserType : " + err.Error())
    }

    return GetUserTypeWithPaper(stub, paper)
}

func GetUserTypeWithPaper (stub shim.ChaincodeStubInterface, paper *Paper) (UserType, error) {
    if CertifyBroker(stub) == nil {
        return ROP, nil
    }

    //EIC
    if CertifyEIC(stub, paper.Organization) == nil {
        return EIC, nil
    }

    //Editor of Paper
    if CertifyEOPWithPaper(stub, paper) == nil {
        return EOP, nil
    }

    //Editor of Organization
    if CertifyEOOWithPaper(stub, paper) == nil {
        return EOO, nil
    }

    //Editor
    if CertifyEditor(stub) == nil {
        return EDITOR, nil
    }

    //Author
    if CertifyAuthorWithPaper(stub, paper) == nil {
        return AUTHOR, nil
    }

    //Reviewer of Paper
    if CertifyROPWithPaperKey(stub, paper.Key) == nil {
        return ROP, nil
    }

    return REVIEWER, nil
}

func CertifyIdentifier (stub shim.ChaincodeStubInterface, identifier string) (error) {
    return cid.AssertAttributeValue(stub, "identifier", identifier)
}

func CertifyBroker (stub shim.ChaincodeStubInterface) (error) {
    return cid.AssertAttributeValue(stub, "userType", "broker")
}

func CertifyEIC (stub shim.ChaincodeStubInterface, organization string) (error) {
    mspKey, err := GetMSPID(stub)
    if err != nil {
        return errors.New("CertifyEIC : " + err.Error())
    }
    if mspKey != organization {
        return errors.New("CertifyEIC : the user is affiliated to " + mspKey + ", not " + organization + ".")
    }

    return cid.AssertAttributeValue(stub, "userType", "eic")
}

func CertifyEditor (stub shim.ChaincodeStubInterface) (error) {
    return cid.AssertAttributeValue(stub, "userType", "editor")
}

func CertifyEOOWithPaper (stub shim.ChaincodeStubInterface, paper *Paper) (error) {
    var organization = paper.Organization

    mspKey, err := GetMSPID(stub)
    if err != nil {
        return errors.New("CertifyEOOWithPaper : " + err.Error())
    }
    if mspKey != organization {
        return errors.New("CertifyEOOWithPaper : the user is affiliated to " + mspKey + ", not " + organization + ".")
    }

    return CertifyEditor(stub)
}

func CertifyEOPWithPaper (stub shim.ChaincodeStubInterface, paper *Paper) (error) {
    var identifier string
    var err error

    identifier, err = GetIdentifier(stub)
    if err != nil {
        return errors.New("CertifyEOPWithPaper : " + err.Error())
    }

    if CertifyIdentifier(stub, paper.EditorID) != nil {
        paperKey, err := CreatePaperKey(stub, paper.Organization, paper.PaperID)
        if err != nil {
            return errors.New("CertifyEOPWithPaper : " + err.Error())
        }
        return errors.New("CertifyEOPWithPaper : " + identifier + " is not editor of " + paperKey + " paper.")
    }

    return nil
}

func CertifyEOPWithPaperKey (stub shim.ChaincodeStubInterface, paperKey string) (error) {
    var paper *Paper
    var err error

    paper, err = GetPaper(stub, paperKey)
    if err != nil {
        return errors.New("CertifyEOPWithPaperKey : " + err.Error())
    }

    return CertifyEOPWithPaper(stub, paper)
}

func CertifyROPWithPaperKey (stub shim.ChaincodeStubInterface, paperKey string) (error) {
    var identifier string
    var RPKey string
    var reviewerKeyByte []byte
    var reviewerKey string
    var reviewerByte []byte
    var reviewer Reviewer
    var err error

    identifier, err = GetIdentifier(stub)
    if err != nil {
        return errors.New("CertifyROPWithPaperKey : " + err.Error())
    }

    RPKey, err = ConvertReviewerKey(stub, paperKey, identifier)
    if err != nil {
        return errors.New("CertifyROPWithPaperKey : " + err.Error())
    }

    reviewerKeyByte, err = InvokeChaincode(stub, "reviewer", "GetPRKey", []string { RPKey })
    if err != nil {
        return errors.New("CertifyROPWithPaperKey : " + err.Error())
    }
    if len(reviewerKeyByte) == 0 {
        return errors.New("CertifyROPWithPaperKey : " + identifier + " is not reviewer of " + paperKey + " paper.")
    }

    reviewerKey = string(reviewerKeyByte)

    reviewerByte, err = InvokeChaincode(stub, "reviewer", "QueryReviewer", []string { reviewerKey })
    if err != nil {
        return errors.New("CertifyROPWithPaperKey : " + err.Error())
    }
    if len(reviewerKeyByte) == 0 {
        return errors.New("CertifyROPWithPaperKey : " + identifier + " is not reviewer of " + paperKey + " paper.")
    }

    err = json.Unmarshal(reviewerByte, &reviewer)
    if err != nil {
        return errors.New("CertifyROPWithPaperKey : " + err.Error())
    }
    if reviewer.Status != "selected" && reviewer.Status != "accept" && reviewer.Status != "submitted" {
        return errors.New("CertifyROPWithPaperKey : " + identifier + " is not selected as a reviewer of " + paperKey + " paper.")
    }

    return nil
}

func CertifyAuthorWithPaper (stub shim.ChaincodeStubInterface, paper *Paper) (error) {
    var err error

    err = CertifyIdentifier(stub, paper.AuthorID)
    if err != nil {
        return errors.New("CertifyAuthorWithPaper : " + err.Error())
    }

    return nil
}

func CertifyAuthorWithPaperKey (stub shim.ChaincodeStubInterface, paperKey string) (error) {
    var paper *Paper
    var err error

    paper, err = GetPaper(stub, paperKey)
    if err != nil {
        return errors.New("CertifyAuthorWithPaperKey : " + err.Error())
    }

    return CertifyAuthorWithPaper(stub, paper)
}

func CheckPaperStatus(stub shim.ChaincodeStubInterface, paperKey string, status string) (error) {
    var paper *Paper
    var err error

    paper, err = GetPaper(stub, paperKey)
    if err != nil {
        return errors.New("CheckPaperStatus : " + err.Error())
    }

    if paper.Status != status {
        return errors.New("CheckPaperStatus : " + paperKey + " paper is in " + paper.Status + ", not in " + status + " status.")
    }

    return nil
}

func CheckReviewEndWithPaperKey(stub shim.ChaincodeStubInterface, paperKey string) (error) {
    var paper *Paper
    var err error

    paper, err = GetPaper(stub, paperKey)
    if err != nil {
        return errors.New("CheckPaperStatus : " + err.Error())
    }

    if paper.Status != "accepted" && paper.Status != "rejected" {
        return errors.New("CheckPaperStatus : " + paperKey + " paper is in review.")
    }

    return nil
}

func CheckContractStatus(stub shim.ChaincodeStubInterface, contractKey string, status string) (error) {
    var paperKey string
    var err error

    paperKey, err = CreatePaperKeyWithContractKey(stub, contractKey)
    if err != nil {
        return errors.New("CheckContractStatus : " + err.Error())
    }

    return CheckPaperStatus(stub, paperKey, status)
}

func CheckReviewerStatus(stub shim.ChaincodeStubInterface, reviewerKey string, status string) (error) {
    var reviewerByte []byte
    var reviewer Reviewer
    var err error

    reviewerByte, err = InvokeChaincode(stub, "reviewer", "QueryReviewer", []string { reviewerKey })
    if err != nil {
        return errors.New("CheckReviewerStatus : " + err.Error())
    }

    err = json.Unmarshal(reviewerByte, &reviewer)
    if err != nil {
        return errors.New("CheckReviewerStatus : " + err.Error())
    }

    if reviewer.Status != status {
        return errors.New("CheckReviewerStatus : " + reviewerKey + " is in " + reviewer.Status + ", not in " + status + " status.")
    }

    return nil
}

