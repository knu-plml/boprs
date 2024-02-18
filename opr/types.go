package main

type UserType int
const (
    EIC = iota
    EOP
    EOO
    EDITOR
    AUTHOR
    ROP
    REVIEWER
    SELF
)

type Paper struct {
    ID             string
    Organization    string
    PaperKey         string
    Title           string
    Abstract        string
    AuthorID        string
    AuthorEmail     string
    EditorID        string
    Round           int
    Status          string
    ContractID     string
}

type PaperFile struct {
    ID             string
    File            string
    Date            string
}

type PaperHistory struct {
    Paper           Paper
    Timestamp       string
}

type Reviewer struct {
    ID             string
    PaperID        string
    ReviewerID      string
    Email           string
    Status          string
}

type Contract struct {
    ID             string
    PaperID        string
    Round           int
    DueDate         string
    CompleteDate    string
}

type Signature struct {
    ID             string
    ContractID     string
    PaperID        string
    ReviewerID     string
    Signature       string
}

type Comment struct {
    ID             string
    ContractID     string
    ReviewerID      string
    Location        []string
    Comment         []string
}

type Report struct {
    ID             string
    ContractID     string
    OverallComment  Comment
    Decision        string
}

type Message struct {
    ID             string
    From            string
    To              string
    PaperID        string
    Round           string
    Message         string
}

type Rating struct {
    ID              string
    ReviewerID      string
    PRID            string
    PaperID         string
    AuthorRating    string
    EditorRating    string
}

type AcceptanceModel struct {
    Organization    string
    OI              string
    OR              string
    OP              string
    ON              string
    OM              string
    OC              string
    OPl             string
}
