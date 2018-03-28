/// 客服系统
// 当前的用户id，默认是所有未读消息的第一个用户id
{/* <script type="text/javascript"></script> */}
var current_uid = -1; 
var ws;
//创建websocket对象
ws = new WebSocket('ws://' + window.location.host + '/custom_service');
window.addEventListener("load", function(evt) {
	console.log('load...');
	//获取所有的未读消息
	 user_list();
	ws.onopen = function(evt) {
		console.log('ws opened');
	}
	ws.onclose = function(evt) {
		console.log('ws closed');
		ws = null;
	}
	ws.onmessage = function(e) {
		console.log('ws onmessage');
		var msg = JSON.parse(e.data);
		answers(msg);
	}
	// 5秒更新一次用户消息
	// setTimeout(user_list,5000); 
	setInterval("user_list()",5000);
});
function judge_update(msg,is_bg) {
	console.log("judge_update" + msg.uid)
	// alert(document.getElementById(msg.uid).innerHTML);
	if (document.getElementById(msg.uid)){
		// alert ("666");
		$("#"+msg.uid+ " .infor").innerHTML = msg.len + "条未读消息";
		return "";
	}
	else {
		if (is_bg){
			return  '<li class="bg" onclick="on_left_li_clicked('+msg.uid+')" id='+msg.uid+'>'+
		'<div class="liLeft"><img src="img/20170926103645_04.jpg"/></div>'+
		'<div class="liRight"><span class="intername">用户:'+msg.uid+'</span>'+
		'<span class="infor">'+msg.len +'条未读消息</span></div></li>'
		}else {
			return '<li onclick="on_left_li_clicked('+msg.uid+')" id='+msg.uid+'>'+
			'<div class="liLeft"><img src="img/20170926103645_04.jpg"/></div>'+
			'<div class="liRight"><span class="intername">用户:'+msg.uid+'</span>'+
		   '<span class="infor">'+msg.len +'条未读消息</span></div></li>'
		}
	}
}
// ajax获取用户列表
function user_list() {
	$.ajax({  
		type: "GET",  
		url: "http://" + window.location.host + "/user_msg_list",  
		contentType: "application/json", //必须有  
		dataType: "json", //表示返回值类型，不必须  
		// data: JSON.stringify({ 'foo': 'foovalue', 'bar': 'barvalue' }),  //相当于 //data: "{'str1':'foovalue', 'str2':'barvalue'}",  
		success: function (msg) {  
			var str = "";
			if (msg.length >= 1  && current_uid == -1){
				// 首页显示的
				message_list(msg[0].uid);
				str += judge_update(msg[0],true)
			} 
			for (var i = 1; i < msg.length; i++){
				str += judge_update(msg[i],false)
			}
			$('.user_list').append(str);
		}  
	});  
}
//ajax 获取消息列表
function message_list (uid) {
	current_uid = uid;
	$.ajax({  
		type: "GET",  
		url: "http://" + window.location.host + "/msg_info?uid=" + uid,  
		contentType: "application/json", //必须有  
		dataType: "json", //表示返回值类型，不必须  
		success: function (msg) { 
			var str='';
			for (var i = 0; i < msg.length; i++){
				// alert(msg[i].uid);  
				str+='<li>'+
				'<div class="nesHead"><img src="img/tou.jpg"/></div>'+
				'<div class="news"><img class="jiao" src="img/jiao.jpg">'+msg[i].message+'</div>'+
			'</li>';
				$('.newsList').append(str);
			}
		
		}  
	});  
}

function on_left_li_clicked(id) {
		//$("#"+id)
		$("#"+id).addClass('bg').siblings().removeClass('bg');
		var intername=$("#"+id).children('.liRight').children('.intername').text();
		$('.headName').text(intername);
		$('.newsList').html('');
		message_list(id);
}
$('.conLeft li').on('click',function(){
		$(this).addClass('bg').siblings().removeClass('bg');
		var intername=$(this).children('.liRight').children('.intername').text();
		$('.headName').text(intername);
		$('.newsList').html('');
		message_list($(this).id);
	});
	$('.sendBtn').on('click',function(){
		var news=$('#dope').val();
		if(news==''){
			alert('不能为空');
		}else{
				//websocket发送消息
		ws.send(
            JSON.stringify({
                uid:1,
				receiver_uid:current_uid,//目前第一版，只支持一个客服，id为1
                message: news 
              }
    	));
		$('#dope').val('');
		var str='';
		str+='<li>'+
		'<div class="answerHead"><img src="img/6.jpg"/></div>'+
		'<div class="answers"><img class="jiao" src="img/20170926103645_03_02.jpg">'+news+'</div>'+
	'</li>';

		$('.newsList').append(str);
		// setTimeout(answers,1000); 
		$('.conLeft').find('li.bg').children('.liRight').children('.infor').text(news);
		$('.RightCont').scrollTop($('.RightCont')[0].scrollHeight );
	
	}
	})
	function answers(msg){
		// var arr=["你好","今天天气很棒啊","你吃饭了吗？","我最美我最美","我是可爱的僵小鱼","你们忍心这样子对我吗？","spring天下无敌，实习工资850","我不管，我最帅，我是你们的小可爱","段友出征，寸草不生","一入段子深似海，从此节操是路人","馒头：嗷","突然想开个车","段子界混的最惨的两个狗：拉斯，普拉达。。。"];
		// var aa=Math.floor((Math.random()*arr.length));
		if (msg.uid == current_uid){
			var answer='<li>'+
				'<div class="nesHead"><img src="img/tou.jpg"/></div>'+
				'<div class="news"><img class="jiao" src="img/jiao.jpg">'+msg.message+'</div>'+
			'</li>';
			$('.newsList').append(answer);	
		}
		$('.RightCont').scrollTop($('.RightCont')[0].scrollHeight );
	}
	// $('.ExP').on('mouseenter',function(){
	// 	$('.emjon').show();
	// })
	// $('.emjon').on('mouseleave',function(){
	// 	$('.emjon').hide();
	// })
	$('.emjon li').on('click',function(){
		var imgSrc=$(this).children('img').attr('src');
		var str="";
		str+='<li>'+
				'<div class="nesHead"><img src="img/6.jpg"/></div>'+
				'<div class="news"><img class="jiao" src="img/20170926103645_03_02.jpg"><img class="Expr" src="'+imgSrc+'"></div>'+
			'</li>';
		$('.newsList').append(str);
		$('.emjon').hide();
		$('.RightCont').scrollTop($('.RightCont')[0].scrollHeight );
	})
