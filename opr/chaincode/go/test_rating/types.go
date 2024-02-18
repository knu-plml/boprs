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
    Key             string
    Organization    string
    PaperID         string
    Title           string
    Abstract        string
    AuthorID        string
    AuthorEmail     string
    EditorID        string
    Round           int
    Status          string
    ContractKey     string
}

type PaperFile struct {
    Key             string
    File            string
    Date            string
}

type PaperHistory struct {
    Paper           Paper
    Timestamp       string
}

type Reviewer struct {
    Key             string
    PaperKey        string
    ReviewerID      string
    Email           string
    Status          string
}

type Contract struct {
    Key             string
    PaperKey        string
    Round           int
    DueDate         string
    CompleteDate    string
}

type Signature struct {
    Key             string
    ContractKey     string
    PaperKey        string
    ReviewerID     string
    Signature       string
}

type Comment struct {
    Key             string
    ContractKey     string
    ReviewerID      string
    Location        []string
    Comment         []string
}

type Report struct {
    Key             string
    ContractKey     string
    OverallComment  Comment
    Decision        string
}

type Message struct {
    Key             string
    From            string
    To              string
    PaperKey        string
    Round           string
    Message         string
}

type Rating struct {
    Key              string
    ReviewerID      string
    PRKey            string
    PaperKey         string
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
