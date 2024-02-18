var express = require('express')
var router = express.Router()
var moment = require('moment')
var path = require('path')

var uploadPaper = require('../opr-test/paper/AddPaper')
var queryCurrentPaper = require('../opr-test/paper/QueryCurrentPaperFile')
var updatePaper = require('../opr-test/paper/UpdatePaper')
var certEditor = require('../opr-test/CertifyEditor')

//파일관련 모듈
var multer = require('multer')

//파일 업로드 모듈
var upload = multer()

// 안쓸것?
router.post('/upload_paper', upload.single('paper'), function (req, res) {
  //req.body.user
  //uploadPaper.main(req.body.user, req.body.org, req.body.author, req.body.email, req.body.pid, req.body.ptitle, req.body.pabt, req.file.buffer.toString('base64'))
  uploadPaper.main('editor@kangwon.ac.kr', req.body.org, req.body.author, req.body.email, req.body.pid, req.body.ptitle, req.body.pabt, req.file.buffer.toString('base64'))
  .then(function() {
    res.redirect('/');
  })
});

router.post('/upload_revision', upload.fields([{ name: 'paper' }, { name: 'note' }]), function (req, res) {
  data = JSON.parse(req.body.data.toString());
  updatePaper.main(req.body.user, data.Key, req.files.paper[0].buffer.toString('base64'), req.files.note[0].buffer.toString('base64'))
  .then(function() {
    res.render('sendForm', {
      url: req.headers.origin + '/login/1',
      method: 'post',
      items: {
        'user': req.body.user
      }
    })
  })
});

router.post('/download/', function (req, res) {
  const paperKey = req.body.paperKey.replace(/�/gi, '\u0000')
  queryCurrentPaper.main(req.body.user, paperKey)
  .then(function(result) {
    result = JSON.parse(result.toString());
    res.contentType('application/vnd.openxmlformats-officedocument.wordprocessingml.document');
    res.setHeader('content-disposition', 'attachment; filename=' + req.body.title + '_' + req.body.round +'.docx');
    res.send(new Buffer(result.File, 'base64'))
  })
})

router.post('/download_revision', function (req, res) {
  const paperKey = req.body.paperKey.replace(/�/gi, '\u0000')
  queryCurrentPaper.main(req.body.user, paperKey)
  .then(function(result) {
    result = JSON.parse(result.toString());
    if (!result.RevisionNote) {
      res.redirect("/error/There is no revision note")
      return
    }
    res.contentType('application/vnd.openxmlformats-officedocument.wordprocessingml.document');
    res.setHeader('content-disposition', 'attachment; filename=' + req.body.title + '_' + req.body.round +'_revision_note.docx');
    res.send(new Buffer(result.RevisionNote, 'base64'))
    //res.send(new Buffer(result.RevisionNote.File, 'base64'))
  })
})

router.post("/efile", function (req, res) {
  res.render('editorSubmit', {
    user: req.body.user
  })
})

router.post("/file", function (req, res) {
  res.render('authorSubmit', {
    user: req.body.user
  })
})

module.exports = router
