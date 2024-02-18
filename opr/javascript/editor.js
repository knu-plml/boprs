var express = require('express') 
var nodemailer = require('nodemailer')
var router = express.Router() 
var path = require('path')

var queryPaper = require('../opr-test/paper/QueryPaperWithEditorID')
var queryReviewer = require('../opr-test/reviewer/QueryReviewerWithPaperKey')
var rejectReviewer = require('../opr-test/reviewer/RejectReviewer')
var stopRecuritNSendBroker = require('../opr-test/broker/SelfRecList-producer')
var createContract = require('../opr-test/contract/CreateContract')
var nextContract = require('../opr-test/contract/NextContract')
var startRound = require('../opr-test/contract/Fulfillment')
var endRound = require('../opr-test/contract/CompleteContract')
var queryComment = require('../opr-test/comment/QueryCommentWithContractKey')
var queryReport = require('../opr-test/report/QueryReport')
var queryContract = require('../opr-test/contract/QueryContract')
var addReportNSendBroker = require('../opr-test/broker/QueryReport-producer')
var queryHistory = require('../opr-test/paper/QueryPaperHistory')
var certEditor = require('../opr-test/CertifyEditor')
var changedStatusNSendBroker = require('../opr-test/broker/ContractStatus-producer')
var getID = require('../opr-test/paper/GetIdentifier')

function numberPad(n, width) {
  n = n + '';
  return n.length >= width ? n : new Array(width - n.length + 1).join('0') + n;
}

function sendEmail(sender, receiver, flag, paper) {
  var subject = ''
  var text = ''
  var transporter = nodemailer.createTransport({
    host: "smtp.gmail.com",
    secure: true,
    auth: {
      //type: "login",
      user: 'opr.testmail@gmail.com',
      pass: 'pl@8441308307'
    }
  })

  if (flag == 1) { // In the case of selection of reviewers.
    subject = '[BOPR] You was selected a reviewer.'
    content = "You was selected a reviewer on the paper '" + paper + "'.\n\nVisit the website and accept to be a reviewer. \n\nhttp://" + req.headers.origin + ":3000"
  }

  var mailOptions = {
    from: sender + ' <opr.testmail@gmail.com>',
    to: receiver,
    subject: subject,
    text: content
  }

  transporter.sendMail(mailOptions, function(error, info){
    if (error) {
      console.log(error);
    }
    else {
      console.log('Email sent: ' + info.response);
    }
  })
}

//게시판 페이징
router.post("/editor_list/:cur", function (req, res) {

  if (!req.body.user){
    res.redirect('/')
    return
  }
  certEditor.main(req.body.user)
  .then(function(result) {
    data = JSON.parse(result.toString());
    if (data){

      //페이지당 게시물 수 : 한 페이지 당 10개 게시물
      var page_size = 10;
      //페이지의 갯수 : 1 ~ 10개 페이지
      var page_list_size = 10;
      //limit 변수
      var no = "";
      //전체 게시물의 숫자
      var totalPageCount = 0;
    
      getID.main(req.body.user)
      .then(function(result) {
        id = result.toString()
        queryPaper.main(req.body.user, id)
        .then(function(result) {
          data = JSON.parse(result.toString());
      
          //전체 게시물의 숫자
          totalPageCount = data.length
      
          //현재 페이지
          var curPage = req.params.cur;
      
          //전체 페이지 갯수
          if (totalPageCount < 0) {
            totalPageCount = 0
          }
      
          var totalPage = Math.ceil(totalPageCount / page_size);// 전체 페이지수
          var totalSet = Math.ceil(totalPage / page_list_size); //전체 세트수
          var curSet = Math.ceil(curPage / page_list_size) // 현재 셋트 번호
          var startPage = ((curSet - 1) * 10) + 1 //현재 세트내 출력될 시작 페이지
          var endPage = (startPage + page_list_size) - 1; //현재 세트내 출력될 마지막 페이지
      
          //현재페이지가 0 보다 작으면
          if (curPage < 0) {
            no = 0
          } else {
            //0보다 크면 limit 함수에 들어갈 첫번째 인자 값 구하기
            no = (curPage - 1) * 10
          }
      
      
          var result2 = {
            "curPage": curPage,
            "page_list_size": page_list_size,
            "page_size": page_size,
            "totalPage": totalPage,
            "totalSet": totalSet,
            "curSet": curSet,
            "startPage": startPage,
            "endPage": endPage
          };
          data = data.slice(no, no + page_size)
          res.render('editorList', {
            data: data,
            user: req.body.user,
            pasing: result2,
            filename: path.join(__dirname)
          })
        })
      })
    }
    else{
      res.redirect("/error/You aren't an editor.")
    }
  })
})

router.post("/edit/:title", function(req, res) {
  data = JSON.parse(req.body.data.toString());
  const paperKey = data.Key.replace(/�/gi, '\u0000')
  queryReviewer.main(req.body.user, paperKey)
  .then(function(result) {
    cand = JSON.parse(result.toString());
    res.render('editorManagement', {
      title: req.params.title,
      user: req.body.user,
      data: data,
      cand: cand
    })
  })
})
router.post("/edit/:title/report/:round", function (req, res) {
  data = JSON.parse(req.body.data.toString());
  var contractKey = data.ContractKey.replace(/�/gi, '\u0000')
  var length = 2
  if (data.Round > 9) {
    length = 3
  }
  contractKey = contractKey.slice(0, -length) + String(req.params.round) + '\u0000'
  var decision = 'under_review or under_decision'
  var cComment = ''

  queryContract.main(req.body.user, contractKey)
  .then(function(result) {
    cont = JSON.parse(result.toString());
  })
  queryComment.main(req.body.user, contractKey)
  .then(function(result) {
    rComment = JSON.parse(result.toString());
    queryReport.main(req.body.user, contractKey)
    .then(function(result2) {
      data2 = JSON.parse(result2.toString());
      if (data2) {
        cComment = data2.OverallComment
        decision = data2.Decision
      }
      res.render('editorReport', {
        title: req.params.title,
        user: req.body.user,
        data: data,
        cont: cont,
        cRound: req.params.round,
        decision: decision,
        rComment: rComment,
        cComment: cComment
      })
    })
  })
})

router.post("/edit/:title/:process", function (req, res) {
  data = JSON.parse(req.body.data.toString());
  const paperKey = data.Key.replace(/�/gi, '\u0000')
  var contractKey = data.ContractKey.replace(/�/gi, '\u0000')

  if (req.params.process == 'send') {
    var loc = new Array
    var comment = new Array
    for (var i = 1; i < req.body.loc.length; i++) {
      if (String(req.body.loc[i]).trim() != '' && String(req.body.comment[i]).trim() != '') {
        loc.push(req.body.loc[i])
        comment.push(req.body.comment[i])
      }
    }
    loc = JSON.stringify(loc)
    comment = JSON.stringify(comment)
  
    addReportNSendBroker.main(req.body.user, data.PaperID, contractKey, data.Round, loc, comment, req.body.decision, brokerAddress)
    .then(function() {
      changedStatusNSendBroker.main(req.body.user, data.PaperID, brokerAddress)
      .then(function() {
        queryReviewer.main(req.body.user, paperKey)
        .then(function(result) {
          cand = JSON.parse(result.toString())
          res.render('sendForm', {
            url: req.headers.origin + '/editor_list/1',
            method: 'post',
            items: {'data': req.body.data,
                    'user': req.body.user,
                    'title': req.params.title,
                    'cand': cand
            }
          })
        })
      })
    })
  }
  else if (req.params.process == 'stop') {
    stopRecuritNSendBroker.main(req.body.user, data.PaperID, brokerAddress)
    .then(function() {
      changedStatusNSendBroker.main(req.body.user, data.PaperID, brokerAddress)
      .then(function() {
        queryReviewer.main(req.body.user, paperKey)
        .then(function(result) {
          cand = JSON.parse(result.toString())
          for (var i in cand) {
            if (cand[i].Status == "candidate"){
              console.log(cand[i])
              //sendEmail(req.body.user, cand[i].Email, 1, req.params.title)
            }
          }
          res.render('sendForm', {
            url: req.headers.origin + '/editor_list/1',
            method: 'post',
            items: {'data': req.body.data,
                    'user': req.body.user,
                    'title': req.params.title,
                    'cand': cand
            }
          })
        })
      })
    })
  }
  else if (req.params.process == 'contract') {
    var time = String(req.body.year) + '-' +  numberPad(req.body.month, 2) + '-' + numberPad(req.body.day, 2) + ' '
    // 심사 시작 전(맨처음)
    if (data.Round == 0){
      createContract.main(req.body.user, paperKey, time)
      .then(function() {
        changedStatusNSendBroker.main(req.body.user, data.PaperID, brokerAddress)
        .then(function() {
          queryReviewer.main(req.body.user, paperKey)
          .then(function(result) {
            cand = JSON.parse(result.toString())
            res.render('sendForm', {
              url: req.headers.origin + '/editor_list/1',
              method: 'post',
              items: {'data': req.body.data,
                      'user': req.body.user,
                      'title': req.params.title,
                      'cand': cand
              }
            })
          })
        })
      })
    }
    //1라운드 이상
    else {
      nextContract.main(req.body.user, contractKey, time)
      .then(function() {
        queryReviewer.main(req.body.user, paperKey)
        .then(function(result) {
          cand = JSON.parse(result.toString())
          res.render('sendForm', {
            url: req.headers.origin + '/editor_list/1',
            method: 'post',
            items: {'data': req.body.data,
                    'user': req.body.user,
                    'title': req.params.title,
                    'cand': cand
            }
          })
        })
      })
    }
  }
  else if (req.params.process == 'round') {
    startRound.main(req.body.user, contractKey)
    .then(function() {
      changedStatusNSendBroker.main(req.body.user, data.PaperID, brokerAddress)
      .then(function() {
        queryReviewer.main(req.body.user, paperKey)
        .then(function(result) {
          cand = JSON.parse(result.toString())
          res.render('sendForm', {
            url: req.headers.origin + '/editor_list/1',
            method: 'post',
            items: {'data': req.body.data,
                    'user': req.body.user,
                    'title': req.params.title,
                    'cand': cand
            }
          })
        })
      })
    })
  }
  else if (req.params.process == 'end') {
    endRound.main(req.body.user, contractKey)
    .then(function() {
      changedStatusNSendBroker.main(req.body.user, data.PaperID, brokerAddress)
      .then(function() {
        queryReviewer.main(req.body.user, paperKey)
        .then(function(result) {
          cand = JSON.parse(result.toString())
          res.render('sendForm', {
            url: req.headers.origin + '/editor_list/1',
            method: 'post',
            items: {'data': req.body.data,
                    'user': req.body.user,
                    'title': req.params.title,
                    'cand': cand
            }
          })
        })
      })
    })
  }
  else if (req.params.process == 'delete') {
    const reviewerKey = req.body.reviewerKey.replace(/�/gi, '\u0000')
    rejectReviewer.main(req.body.user, reviewerKey)
    .then(function() {
      queryReviewer.main(req.body.user, paperKey)
      .then(function(result) {
        cand = JSON.parse(result.toString())
        res.render('sendForm', {
          url: req.headers.origin + '/editor_list/1',
          method: 'post',
          items: {'data': req.body.data,
                  'user': req.body.user,
                  'title': req.params.title,
                  'cand': cand
          }
        })
      })
    })
  }
  else if (req.params.process == 'history') {
    queryHistory.main(req.body.user, paperKey)
    .then(function(result) {
      hist = JSON.parse(result.toString());
      res.render('editorHistory', {
        title: req.params.title,
        data: data,
        hist: hist,
        user: req.body.user
      })
    })
  }
  else if (req.params.process == 'update') {
    res.render('editorRevise', {
      title: req.params.title,
      data: data,
      user: req.body.user
    })
  }
  else if (req.params.process == 'rating') {
    queryReviewer.main(req.body.user, paperKey)
    .then(function(result) {
      reviewers = JSON.parse(result.toString());
      var selected_reviewers = []
      for (i = 0; i < reviewers.length; i++) {
        if (reviewers[i]['Status'] == 'submitted' || reviewers[i]['Status'] == 'accept') {
          selected_reviewers.push(reviewers[i])
        }
      }
      res.render('editorRating', {
        title: req.params.title,
        user: req.body.user,
        reviewers: selected_reviewers,
        data: data
      })
    })
  }
})

module.exports = router
