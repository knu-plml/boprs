var express = require('express')
var router = express.Router()
var path = require('path')

var queryPaper = require('../opr-test/paper/QueryPaperWithAuthorID')
var queryComment = require('../opr-test/comment/QueryCommentWithContractKey')
var queryReviewer = require('../opr-test/reviewer/QueryReviewerWithPaperKey')
var queryContract = require('../opr-test/contract/QueryContract')
var queryReport = require('../opr-test/report/QueryReport')
var queryHistory = require('../opr-test/paper/QueryPaperHistory')
var getID = require('../opr-test/paper/GetIdentifier')
var queryPaperComment = require('../opr-test/message/QueryMessageWithContractKey')
var addPaperComment = require('../opr-test/message/AddMessage')

router.post('/author_list/:cur', function (req, res) {

  //페이지당 게시물 수 : 한 페이지 당 10개 게시물
  var page_size = 10;
  //페이지의 갯수 : 1 ~ 10개 페이지
  var page_list_size = 10;
  //limit 변수
  var no = "";
  //전체 게시물의 숫자
  var totalPageCount = 0;
  if (!req.body.user){
    res.redirect('/')
    return
  }

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
      result = data.slice(no, page_size)
      res.render('authorList', {
        user: req.body.user,
        data: result,
        pasing: result2,
        filename: path.join(__dirname)
      })
    })
  })
})

router.post('/author/:title/:round', function(req, res) {
  const data = JSON.parse(req.body.data.toString());
  const paperKey = data.Key.replace(/�/gi, '\u0000')
  var contractKey = data.ContractKey.replace(/�/gi, '\u0000')

  if (req.params.round == 'history') {
    queryHistory.main(req.body.user, paperKey)
    .then(function(result) {
      hist = JSON.parse(result.toString());
      console.log(hist)
      res.render('authorHistory', {
        title: req.params.title,
        user: req.body.user,
        hist: hist,
        data: data
      })
    })
  }
  else if (req.params.round == 'rating') {
    queryReviewer.main(req.body.user, paperKey)
    .then(function(result) {
      reviewers = JSON.parse(result.toString());
      var selected_reviewers = []
        for (i = 0; i < reviewers.length; i++) {
          if (reviewers[i]['Status'] == 'submitted' || reviewers[i]['Status'] == 'accept') {
            selected_reviewers.push(reviewers[i])
          }
      }
      res.render('authorRating', {
        title: req.params.title,
        user: req.body.user,
        reviewers: selected_reviewers,
        data: data
      })
    })
  }
  else if (req.params.round == 'update') {
    res.render('authorRevise', {
      title: req.params.title,
      user: req.body.user,
      data: data
    })
  }
  else {
    var length = 2
    if (data.Round > 9) {
      length = 3
    }
    if (contractKey == '') {
      contractKey = paperKey
    }
    else {
      contractKey = contractKey.slice(0, -length) + String(req.params.round) + '\u0000'
    }
    var decision = 'under_review or under_decision'
    var cComment = ''
  
    queryContract.main(req.body.user, contractKey)
    .then(function(result) {
      cont = JSON.parse(result.toString());
      if (cont == 0) {
        cont = ''
      }
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
        queryPaperComment.main(req.body.user, contractKey)
        .then(function(result3) {
          data3 = JSON.parse(result3.toString());
           if (data3.length == 0) {
             data3 = null
           }

          res.render('authorComment', {
            title: req.params.title,
            user: req.body.user,
            data: data,
            cont: cont,
            pComment: data3,
            decision: decision,
            cRound: req.params.round,
            rComment: rComment,
            cComment: cComment,
          })
        })
      })
    })
  }
})



module.exports = router
