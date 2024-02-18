var express = require('express')
var router = express.Router()
var path = require('path')

var existUser = require('../opr-test/ExistUser')
var queryAcc = require('../opr-test/acceptanceModel/QueryAcceptanceModel')
var updateAcc = require('../opr-test/acceptanceModel/UpdateAcceptanceModel')
var certEditor = require('../opr-test/CertifyEditor')
var certEic = require('../opr-test/CertifyEIC')

router.post('/eicAM', function (req, res) {
  if(!req.body.user) {
    res.redirect('/')
    return
  }
  existUser.main(req.body.user)
  .then(function(result) {
    if (result) {
      certEditor.main(req.body.user)
      .then(function(editor) {
        editor = JSON.parse(editor.toString());
        certEic.main(req.body.user)
        .then(function(eic) {
          eic = JSON.parse(eic.toString());
          if (editor || eic ) {
            page='eicAcceptanceModel';
          }
          else {
            page='userAcceptanceModel';
          }
          queryAcc.main(req.body.user)
          .then(function(result) {
            data = JSON.parse(result.toString())
            res.render(page, {
              user: req.body.user,
              oc: data.OC,
              oi: data.OI,
              om: data.OM,
              on: data.ON,
              op: data.OP,
              or: data.OR
            })
          })
        })
      })
    }
    else {
      result.render('signup', {
        orcid: req.body.user
      })
    }
  })
})

router.post('/eicAM/update', function (req, res) {
  updateAcc.main(req.body.user, req.body.OI, req.body.OR, req.body.OP, req.body.ON, req.body.OM, req.body.OC, 0)
  .then(function() {
    res.render('sendForm', {
      url: req.headers.origin + '/eicAM',
      method: 'post',
      items: {'user': req.body.user
      }
    })
  })
  .catch(function(error) {
    res.redirect("/error/You can't change the acceptance model.")
  })
})

module.exports = router
