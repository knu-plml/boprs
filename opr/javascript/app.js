//모듈 임포트
var express = require('express')
var multiparty = require('connect-multiparty')
var app = express()
var morgan = require('morgan') //로그 모듈 임포트

global.brokerAddress = process.argv.slice(2)

//미들웨어 설정
app.use(morgan('short')) //로그 미들웨어
app.use(express.static('../papers')) //기본 파일 폴더 위치 설정
//app.use(multiparty())
app.use(express.urlencoded({extend:true}))
app.use(express.json())
app.set('view engine', 'pug')
app.set('views', '../views')

//라우트로 분리시켜주기
var fileRouter = require('./file.js')
var reviewRouter = require('./author.js')
var authorRouter = require('./reviewer.js')
var editorRouter = require('./editor.js')
var userRouter = require('./user.js')
var AMRouter = require('./acceptanceModel.js')

app.use(fileRouter)
app.use(reviewRouter)
app.use(authorRouter)
app.use(editorRouter)
app.use(userRouter)
app.use(AMRouter)

//서버 가동
app.listen(3000,function(){
console.log("server on")
})
