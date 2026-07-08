/*
 * AJO Pay 桌面頁互動
 * 1. 保留原型互動與路由切換。
 * 2. 移除重複函式與桌面版不使用的隱藏流程。
 * 3. 維持原始 DOM id 與 onclick 呼叫，方便逐步遷移。
 */
// 1. 切換 AJO Pay 內部頁面
function dGo(i,el){
  document.querySelectorAll('[id^="dpg"]').forEach(p=>p.classList.remove('on'));
  document.querySelectorAll('.d-tab').forEach(t=>t.classList.remove('on'));
  var pg=document.getElementById('dpg'+i);
  if(pg) pg.classList.add('on');
  if(el) el.classList.add('on');
  window.scrollTo(0,0);
}

// 2. 回到 AJO Pay 首頁或付款頁
function setAjoPayPage(i){
  var pageIndex=i===1?1:0;
  dGo(pageIndex,null);
}

// 3. 從 AJO Pay 首頁進入付款頁
function startAjoPayPayment(){
  setAjoPayPage(1);
}

// 4. 選擇付款項目
function dSelPayItem(el){
  el.classList.toggle('sel');
  dUpdatePayTotal();
}

// 5. 選擇付款方式
function dSelM(el){
  document.querySelectorAll('.d-method-opt').forEach(o=>{o.classList.remove('sel');o.querySelector('.d-radio').innerHTML='';});
  el.classList.add('sel');el.querySelector('.d-radio').innerHTML='<div class="d-rdot"></div>';
  dUpdatePayTotal();
}

// 6. 更新付款步驟狀態
function dUpdatePaymentSteps(total,hasMethod){
  var steps=document.querySelectorAll('#dpg1 .d-payment-step');
  steps.forEach(function(step,index){
    step.classList.toggle('is-active',index===0||total>0&&index===1||total>0&&hasMethod&&index===2);
  });
}

// 7. 更新付款金額與確認按鈕
function dUpdatePayTotal(){
  var total=0;
  document.querySelectorAll('#dpg1 .d-pay-item.selectable.sel').forEach(function(item){
    total+=Number(item.dataset.amount||0);
  });
  var amount='HK$'+total.toLocaleString('en-US',{minimumFractionDigits:2,maximumFractionDigits:2});
  var totalEl=document.querySelector('#dpg1 .d-pay-total-amount');
  var coinEl=document.querySelector('#dpg1 .d-coin-estimate');
  var btn=document.querySelector('#dpg1 .d-confirm-btn');
  var hasMethod=!!document.querySelector('#dpg1 .d-method-opt.sel');
  if(totalEl)totalEl.textContent=amount;
  if(coinEl)coinEl.textContent=Math.floor(total/100)+' AJO Coins';
  dUpdatePaymentSteps(total,hasMethod);
  if(btn){
    btn.disabled=!(total>0&&hasMethod);
    btn.textContent=total>0?(hasMethod?'確認支付 '+amount:'請選擇付款方式'):'請先選擇付款項目';
  }
}

// 8. 顯示支付成功彈窗
function showModal(){
  var modal=document.getElementById('modal');
  var modalWeb=document.getElementById('modalWeb');
  if(!modal||!modalWeb)return;
  modal.classList.add('show');
  modalWeb.style.display='block';
  modal.style.alignItems='center';
}

// 9. 關閉支付成功彈窗
function closeModal(){
  var modal=document.getElementById('modal');
  if(modal)modal.classList.remove('show');
}
