var express = require('express')
var router = express.Router()
var path = require('path')
var axios = require('axios')

var existUser = require('../opr-test/ExistUser')
var queryPaper = require('../opr-test/paper/QueryAllPaper')
var registerUser = require('../opr-test/registerUser')
var queryReviewer = require('../opr-test/reviewer/QueryReviewerWithPaperKey')
var authorRatingNSendBroker = require('../opr-test/broker/author-QueryReviewerReput-producer')
var editorRatingNSendBroker = require('../opr-test/broker/editor-QueryReviewerReput-producer')
var queryComment = require('../opr-test/comment/QueryCommentWithContractKey')
var queryReport = require('../opr-test/report/QueryReport')
var queryPaperComment = require('../opr-test/message/QueryMessageWithContractKey')
var addPaperComment = require('../opr-test/message/AddMessage')
var certEditor = require('../opr-test/CertifyEditor')
var certEic = require('../opr-test/CertifyEIC')


router.get('/', function (req, res) {
  if(req.query.code){
    axios.post('https://orcid.org/oauth/token?client_id=YOUR_CLIENT_ID&client_secret=YOUR_CLIENT_SECRET&grant_type=authorization_code&redirect_uri=YOUR_URI:3000&code=' + req.query.code)
    .then(function(response) {
      console.log(req)
      res.render('sendForm', {
        url: 'http://' + req.headers.host + '/login/1',
        method: 'post',
        items: {'user': response.data.orcid,
        }
      })
    })
    .catch(function(error) {
      console.log(error)
    })
  }
  else {
    res.render('signin')
  }
})

router.get('/signup', function(req, res) {
  res.render('signup')
})

router.post('/registerUser', function(req, res) {
  //orcid 확인
  var orcid
  if (!req.body.orcid) {
    orcid = req.body.email
  }
  else {
    orcid = req.body.orcid
  }
  registerUser.main(orcid, req.body.email)
  .then(function(){
    res.render('sendForm', {
      url: req.headers.origin + '/login/1',
      method: 'post',
      items: {'user': orcid,
      }
    })
  })
})

router.get('/error/:err', function (req, res) {
  res.render('error',{
    data: req.params.err
  })
})

router.get('/logout', function (req, res) { 
  res.redirect('/')
})

router.post('/login/:cur', function (req, res) {
  if (!req.body.user) {
    res.redirect('/')
    return
  }
  existUser.main(req.body.user)
  .then(function(result) {
    if (result) {
      //페이지당 게시물 수 : 한 페이지 당 10개 게시물
      var page_size = 10;
      //페이지의 갯수 : 1 ~ 10개 페이지
      var page_list_size = 10;
      //limit 변수
      var no = "";
      //전체 게시물의 숫자
      var totalPageCount = 0;

      queryPaper.main(req.body.user)
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

        result = data.slice(no, no + page_size)
        certEditor.main(req.body.user)
        .then(function(editor) {
          editor = JSON.parse(editor.toString());
          certEic.main(req.body.user)
          .then(function(eic) {
            eic = JSON.parse(eic.toString());
            if (editor || eic ) {
              page='editorList1';
            }
            else {
              page='userList';
            }
            res.render(page, {
              user: req.body.user,
              data: result,
              pasing: result2,
              filename: path.join(__dirname)
            })
          })
        })
      })
    }
    else {
      res.render('signup', {
        orcid: req.body.user
      })
    }
  })
})

router.post('/login/search/:cur/', function (req, res) {
  var names = req.body.pName.split(' ')

  existUser.main(req.body.user)
  .then(function(result) {
    if (result) {
      var page_size = 10;
      var page_list_size = 10;
      var no = "";
      var totalPageCount = 0;

      queryPaper.main(req.body.user)
      .then(function(result) {
        rawdata = JSON.parse(result.toString());
        var data = []
        var idx = []

        for(i = 0; i < names.length; i++){
          for(j = 0; j < rawdata.length; j++){
            if(rawdata[j].Title.toLowerCase().indexOf(names[i]) != -1){
              var overlap = false
              for(k = 0; k < idx.length; k++) {
                if(j == idx[k]) {
                  overlap = true
                  break
                }
              }
              if(!overlap){
                idx.push(j)
              }
            }
          }
        }

        for(i = 0; i < idx.length; i++) {
          data.push(rawdata[i])
        }
      
        totalPageCount = data.length
    
        var curPage = req.params.cur;
    
        if (totalPageCount < 0) {
          totalPageCount = 0
        }
    
        var totalPage = Math.ceil(totalPageCount / page_size);
        var totalSet = Math.ceil(totalPage / page_list_size);
        var curSet = Math.ceil(curPage / page_list_size)
        var startPage = ((curSet - 1) * 10) + 1 
        var endPage = (startPage + page_list_size) - 1;
    
        if (curPage < 0) {
          no = 0
        } else {
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

        result = data.slice(no, no + page_size)
        console.log(result)
        res.render('list', {
          user: req.body.user,
          data: result,
          pasing: result2,
          filename: path.join(__dirname)
        })
      })
    }
    else {
      res.redirect('/')
    }
  })
})

router.post('/login/paper/:title/:round', function (req, res) {
  res.render('sendForm', {
    url: req.headers.origin + '/paper/' + req.params.title + '/' + req.params.round,
    method: 'post',
    items: {data: req.body.data,
            user: req.body.user
    }
  })
})

router.post('/:user/:title/:round/comment', function (req, res) {
  data = JSON.parse(req.body.data.toString());
  var contractKey = data.ContractKey.replace(/�/gi, '\u0000')
  var length = 2
  if (data.Round > 9) {
    length = 3
  }
  contractKey = contractKey.slice(0, -length) + String(req.params.round) + '\u0000'
  addPaperComment.main(req.body.user, contractKey, req.body.comment)
  .then(function() {
    if(req.params.user == 'paper') {
      res.render('sendForm', {
        url: req.headers.origin + '/' + req.params.user + '/' + req.params.title + '/' + req.params.round,
        method: 'post',
        items: {data: req.body.data,
                user: req.body.user
        }
      })
    }
    else if(req.params.user == 'reviewer') {
      res.render('sendForm', {
        url: req.headers.origin + '/' + req.params.user + '/' + req.params.title + '/contract',
        method: 'post',
        items: {data: req.body.data,
                user: req.body.user
        }
      })
    }
    else if(req.params.user == 'author') {
      res.render('sendForm', {
        url: req.headers.origin + '/' + req.params.user + '/' + req.params.title + '/' + req.params.round,
        method: 'post',
        items: {data: req.body.data,
                user: req.body.user
        }
      })
    }
  })
  .catch(function(error) {
    res.redirect("/error/You can't leave a message")
  })
})

router.post('/paper/:title/:round', function (req, res) {
  data = JSON.parse(req.body.data.toString());

  if (req.params.round == 0){
    certEditor.main(req.body.user)
    .then(function(editor) {
      editor = JSON.parse(editor.toString());
      certEic.main(req.body.user)
      .then(function(eic) {
        eic = JSON.parse(eic.toString());
        if (editor || eic ) {
          page='editorPaper';
        }
        else {
          page='userPaper';
        }
        res.render(page, {
          title: req.params.title,
          user: req.body.user,
          data: data,
          cRound: req.params.round
        })
      })
    })
  }
  else {
    var contractKey = data.ContractKey.replace(/�/gi, '\u0000')
    var length = 2
    if (data.Round > 9) {
      length = 3
    }
    contractKey = contractKey.slice(0, -length) + String(req.params.round) + '\u0000'
    var decision = 'under_review'
    var cComment = ''
  
    queryComment.main(req.body.user, contractKey)
    .then(function(result) {
      rComment = JSON.parse(result.toString());
      queryReport.main(req.body.user, contractKey)
      .then(function(result2) {
        data2 = JSON.parse(result2.toString());
        if (data2){
          cComment = data2.OverallComment
          decision = data2.Decision
        }
        queryPaperComment.main(req.body.user, contractKey)
        .then(function(result3) {
          data3 = JSON.parse(result3.toString());
          if (data3.length == 0) {
            data3 = null
          }
          
          certEditor.main(req.body.user)
          .then(function(editor) {
            editor = JSON.parse(editor.toString());
            certEic.main(req.body.user)
            .then(function(eic) {
              eic = JSON.parse(eic.toString());
              if (editor || eic ) {
                page='editorPaper';
              }
              else {
                page='userPaper';
              }
              res.render(page, {
                title: req.params.title,
                user: req.body.user,
                data: data,
                pComment: data3,
                cRound: req.params.round,
                decision : decision,
                rComment : rComment,
                cComment : cComment,
              })
            })
          })
        })
      })
    })
  }
})

router.post('/rating/:user', function (req, res) {
  data = JSON.parse(req.body.data.toString());
  const paperKey = data.Key.replace(/�/gi, '\u0000')
  queryReviewer.main(req.body.user, paperKey)
  .then(function(result) {
    reviewers = JSON.parse(result.toString());
    var selected_reviewers = []
    for (i = 0; i < reviewers.length; i++) {
      if (reviewers[i]['Status'] == 'submitted' || reviewers[i]['Status'] == 'accept') {
        selected_reviewers.push(reviewers[i])
      }
    }
    var rateList = []
    if (req.params.user == 1){
      var idxList = []
      for(i = 0; i < selected_reviewers.length; i++) {
        idxList.push(selected_reviewers[i].Key.split('\u0000')[selected_reviewers[i].Key.split('\u0000').length - 2])
        rateList.push(req.body[selected_reviewers[i].Key.replace(/\u0000/gi, '�')])
      }
      //authorRating.main(req.body.user, paperKey, JSON.stringify(idxList), JSON.stringify(rateList))
      authorRatingNSendBroker.main(req.body.user, data.PaperID, JSON.stringify(idxList), JSON.stringify(rateList), brokerAddress)
      .then(function() {
        res.render('sendForm', {
          url: req.headers.origin + '/author_list/1',
          method: 'post',
          items: {'data': req.body.data
          }
        })
      })
    }
    else{
      var revList = []
      for(i = 0; i < selected_reviewers.length; i++) {
        revList.push(selected_reviewers[i].ReviewerID)
        rateList.push(req.body[selected_reviewers[i].ReviewerID])
      }
      //editorRating.main(req.body.user, paperKey, JSON.stringify(revList), JSON.stringify(rateList))
      editorRatingNSendBroker.main(req.body.user, data.PaperID, JSON.stringify(revList), JSON.stringify(rateList), brokerAddress)
      .then(function() {
        res.render('sendForm', {
          url: req.headers.origin + '/editor_list/1',
          method: 'post',
          items: {'data': req.body.data
          }
        })
      })
    }
  })
})

module.exports = router
