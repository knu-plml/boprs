var express = require('express')
var router = express.Router()
var path = require('path')

var reviewersPaper = require('../opr-test/reviewer/QueryReviewPapers')
var queryReviewer = require('../opr-test/reviewer/QueryReviewerWithReviewerID')
var queryPaper = require('../opr-test/paper/QueryPaper')
var addComment = require('../opr-test/comment/AddComment')
var queryContract = require('../opr-test/contract/QueryContract')
var queryComment = require('../opr-test/comment/QueryCommentWithContractKey')
var signContract = require('../opr-test/contract/SignContract')
var declineReview = require('../opr-test/reviewer/DeclineReview')
var register = require('../opr-test/reviewer/RegisterCandidate')
var getID = require('../opr-test/paper/GetIdentifier')
var queryPaperComment = require('../opr-test/message/QueryMessageWithContractKey')
var addPaperComment = require('../opr-test/message/AddMessage')

router.post('/reviewer_list/:cur', function (req, res) {

  if (!req.body.user){
    res.redirect('/')
    return
  }

  //페이지당 게시물 수 : 한 페이지 당 10개 게시물
  var page_size = 10;
  //페이지의 갯수 : 1 ~ 10개 페이지
  var page_list_size = 10;
  //limit 변수
  var no = "";
  //전체 게시물의 숫자
  var totalPageCount = 0;

  var result2 = {};

  getID.main(req.body.user)
  .then(function(result) {
    id = result.toString()
    reviewersPaper.main(req.body.user, id)
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
  
      result2 = {
        "curPage": curPage,
        "page_list_size": page_list_size,
        "page_size": page_size,
        "totalPage": totalPage,
        "totalSet": totalSet,
        "curSet": curSet,
        "startPage": startPage,
        "endPage": endPage
      };
      queryReviewer.main(req.body.user, id)
      .then(function(result) {
        data2 = JSON.parse(result.toString());
        data2 = data2.slice(no, page_size)
  
        for(i = 0; i < data.length; i++) {
          for(j = 0; j < data2.length; j++) {
            if(data[i].Key == data2[j].PaperKey) {
              data[i].Reviewer = data2[j].Status
              break
            }
          }
        }
  
        data = data.slice(no, no +  page_size)
        console.log(data)
        res.render('reviewerList', {
          data: data,
          user: req.body.user,
          pasing: result2,
          filename: path.join(__dirname)
        })
      })
    })
  })
})

router.post('/apply_reviewer/:title', function (req, res) {
  const paperKey = req.body.paperKey.replace(/�/gi, '\u0000') 
  register.main(req.body.user, paperKey)
  .then(function() {
    res.render('sendForm', {
	    url: req.headers.orgin + ':3000/login/1',
      /*'http://39.115.145.90:3000/login/1', */
      method: 'post',
      items: {'user': req.body.user,
      }
    })
  })
})

router.post('/login/apply_reviewer/:title', function (req, res) {
  const paperKey = req.body.paperKey.replace(/�/gi, '\u0000') 
  res.render('sendForm', {
    url: req.headers.origin + '/apply_reviewer/' + req.params.title,
    method: 'post',
    items: {'user': req.body.user,
            'paperKey': paperKey
    }
  })
  /*
  register.main(req.body.user, paperKey)
  .then(function() {
    res.render('sendForm', {
      url: 'http://39.115.145.90:3000/login/1',
      method: 'post',
      items: {'user': req.body.user,
      }
    })
  })
  */
})

router.post('/reviewer/:title/contract', function (req, res) {
  if (req.body.data) {
    data = JSON.parse(req.body.data.toString());
    const contractKey = data.ContractKey.replace(/�/gi, '\u0000') 
    queryPaperComment.main(req.body.user, contractKey)
    .then(function(result2) {
      data2 = JSON.parse(result2.toString())
      if (data2.length == 0) {
        data2 = null
      }
      res.render('reviewerContract', {
        title: req.params.title,
        data: data,
        pComment: data2,
        user: req.body.user
      })
    })
    return
  }
  if (!req.body.ckey) {
    res.redirect("/error/Contract wasn't created.")
  }

  const contractKey = req.body.ckey.replace(/�/gi, '\u0000') 
  queryContract.main(req.body.user, contractKey)
  .then(function(result) {
    if (!result) {
      res.redirect("/error/You aren't a reviewer of this paper.")
    }
    else {
      data = JSON.parse(result.toString());
      data.ContractKey = contractKey
      queryPaperComment.main(req.body.user, contractKey)
      .then(function(result2) {
        data2 = JSON.parse(result2.toString())
        if (data2 == []) {
          data2 = null
        }
        res.render('reviewerContract', {
          title: req.params.title,
          data: data,
          pComment: data2,
          user: req.body.user
        })
      })
    }
  })
})

router.post('/reviewer/:title/audit', function (req, res) {
  data = JSON.parse(req.body.data.toString());
  const contractKey = data.Key.replace(/�/gi, '\u0000')
  queryPaper.main(req.body.user, data.PaperKey)
  .then(function(result) {
    result = JSON.parse(result.toString());
    if (!(result.Status == 'under_review')) {
      res.redirect("/error/This paper isn't under_review")
    }
    else  {
      queryComment.main(req.body.user, contractKey)
      .then(function(result) {
        Comment = JSON.parse(result.toString())
        res.render('reviewerAudit', {
          title: req.params.title,
          Comment: Comment[0],
          data: data,
          user: req.body.user
        })
      })
    }
  })
})

router.post('/reviewer/:title/send', function (req, res) {
  
  data = JSON.parse(req.body.data.toString());
  const contractKey = data.Key.replace(/�/gi, '\u0000')

  var loc = new Array
  var comment = new Array
  for (var i = 1; i < req.body.loc.length; i++) {
    if (String(req.body.loc[i]).trim() != '' && String(req.body.comment[i]).trim() != '') {
      loc.push(req.body.loc[i])
      comment.push(req.body.comment[i])
    }
  }
  Comment = {'Comment': comment, 'Location': loc}
  loc = JSON.stringify(loc)
  comment = JSON.stringify(comment)
  addComment.main(req.body.user, contractKey, loc, comment)
  .then(function() {
    res.render('reviewerAudit', {
      title: req.params.title,
      data: data,
      user: req.body.user
    })
  })
})

router.post('/reviewer/:title/apply', function (req, res) {
  data = JSON.parse(req.body.data.toString());
  const contractKey = data.Key.replace(/�/gi, '\u0000')
  const paperKey = data.PaperKey.replace(/�/gi, '\u0000')
  if (req.body.reviewer == 'Accept') {
    signContract.main(req.body.user, contractKey)
    .then(function() {
      res.render('sendForm', {
        url: req.headers.origin + '/reviewer_list/1',
        method: 'post',
        items: {'data': req.body.data,
                'user': req.body.user
        }
      })
    })
    .catch(function(error) {
      res.redirect("/error/You have already pressed Accept or Reject button.")
    })
  }
  else if (req.body.reviewer == 'Reject') {
    declineReview.main(req.body.user, paperKey)
    .then(function() {
      res.render('sendForm', {
        url: req.headers.origin + '/reviewer_list/1',
        method: 'post',
        items: {'data': req.body.data,
                'user': req.body.user
        }
      })
    })
    .catch(function(error) {
      res.redirect("/error/You have already pressed Accept or Reject button.")
    })
  }
})

module.exports = router
