// UI scripts unificados (module)
// - Nav drawer
// - Carousel home
// - Modal "Cómo comprar"
// - Products drawer/sheet + load more
// - SW registration (idle)

// Nav drawer
(function(){
  const btn=document.querySelector('.nav-toggle');
  const body=document.body;
  const backdrop=document.querySelector('.nav-backdrop');
  if(!btn)return;
  function close(){body.classList.remove('nav-open');btn.setAttribute('aria-expanded','false');}
  btn.addEventListener('click',()=>{const open=body.classList.toggle('nav-open');btn.setAttribute('aria-expanded',open);});
  backdrop&&backdrop.addEventListener('click',close);
  window.addEventListener('keydown',e=>{if(e.key==='Escape' && body.classList.contains('nav-open')) close();});
})();

// Home carousel + modal
(function(){
  const slides=[...document.querySelectorAll('#heroCarousel .hc-slide')];
  const dotsC=document.getElementById('hcDots');
  if(slides.length && dotsC){
    slides.forEach((_,i)=>{const d=document.createElement('div');d.className='dot'+(i===0?' active':'');d.dataset.i=i;d.onclick=()=>go(i,true);dotsC.appendChild(d);});
    let idx=0,timer=null;function go(n,manual){slides[idx].classList.remove('active');dotsC.children[idx].classList.remove('active');idx=(n+slides.length)%slides.length;slides[idx].classList.add('active');dotsC.children[idx].classList.add('active');if(manual){restart();}}
    function next(){go(idx+1);}function start(){timer=setInterval(next,3000);}function restart(){clearInterval(timer);start();}start();
    let startX=0,isSwiping=false;const carousel=document.getElementById('heroCarousel');
    if(carousel){
      carousel.addEventListener('touchstart',e=>{startX=e.touches[0].clientX;isSwiping=true;},{passive:true});
      carousel.addEventListener('touchmove',e=>{if(!isSwiping)return;const dx=e.touches[0].clientX-startX;if(Math.abs(dx)>10){e.preventDefault();}},{passive:false});
      carousel.addEventListener('touchend',e=>{if(!isSwiping)return;isSwiping=false;const dx=e.changedTouches[0].clientX-startX;if(Math.abs(dx)>50){go(dx>0?idx-1:idx+1,true);}},{passive:true});
    }
  }
  const btn=document.getElementById('btnHowBuy');
  const bd=document.getElementById('howBuyBackdrop');
  const modal=bd?bd.querySelector('.modal'):null;
  const closeBtn=document.getElementById('howBuyClose');
  const okBtn=document.getElementById('howBuyOk');
  function open(){ if(!bd||!modal) return; bd.hidden=false; bd.classList.add('show'); modal.classList.add('show'); document.body.style.overflow='hidden'; closeBtn&&closeBtn.focus(); }
  function close(){ if(!bd||!modal) return; bd.classList.remove('show'); modal.classList.remove('show'); document.body.style.overflow=''; setTimeout(()=>{bd.hidden=true;},140); btn&&btn.focus(); }
  btn&&btn.addEventListener('click',open);
  closeBtn&&closeBtn.addEventListener('click',close);
  okBtn&&okBtn.addEventListener('click',close);
  bd&&bd.addEventListener('click',e=>{if(e.target===bd) close();});
  document.addEventListener('keydown',e=>{if(e.key==='Escape'&&!bd?.hidden) close();});
})();

// Products drawer/sheet + chips + load more
(function(){
  const filterBtn=document.querySelector('.btn-filter');
  const sortBtn=document.querySelector('.btn-sort');
  const drawer=document.getElementById('filterDrawer');
  const drawerClose=drawer?drawer.querySelector('.drawer-close'):null;
  const panel=drawer?drawer.querySelector('.drawer-panel'):null;
  const sheet=document.getElementById('sortSheet');
  const form=document.getElementById('filtersForm');
  const sortInput=document.getElementById('sortInput');

  function openDrawer(){
    if(!drawer) return;
    drawer.hidden=false;
    drawer.setAttribute('aria-hidden','false');
    filterBtn&&filterBtn.setAttribute('aria-expanded','true');
    document.body.classList.add('drawer-open');
    const first=panel&&panel.querySelector('input, select, button');
    if(first) first.focus();
  }
  function closeDrawer(){
    if(!drawer) return;
    drawer.hidden=true;
    drawer.setAttribute('aria-hidden','true');
    filterBtn&&filterBtn.setAttribute('aria-expanded','false');
    document.body.classList.remove('drawer-open');
  }
  function openSheet(){ if(!sheet) return; sheet.hidden=false; sortBtn&&sortBtn.setAttribute('aria-expanded','true'); }
  function closeSheet(){ if(!sheet) return; sheet.hidden=true;  sortBtn&&sortBtn.setAttribute('aria-expanded','false'); }

  filterBtn&&filterBtn.addEventListener('click',openDrawer);
  drawerClose&&drawerClose.addEventListener('click',closeDrawer);
  sortBtn&&sortBtn.addEventListener('click',()=>{ if(sheet&&sheet.hidden){ openSheet(); } else { closeSheet(); } });
  window.addEventListener('keydown',(e)=>{
    if(e.key==='Escape'){
      if(drawer&&!drawer.hidden) closeDrawer();
      if(sheet&&!sheet.hidden) closeSheet();
    }
  });
  if(drawer){ drawer.addEventListener('click',(e)=>{ if(e.target===drawer) closeDrawer(); }); }

  // Ordenar: setear hidden input y enviar el form
  document.querySelectorAll('#sortSheet [data-sort]').forEach(btn=>{
    btn.addEventListener('click',()=>{
      if(!sortInput || !form) return;
      sortInput.value=btn.getAttribute('data-sort')||'';
      form.submit();
    });
  });

  // Chips limpiar (si existen)
  document.querySelectorAll('.chip[data-clear]').forEach(btn=>{
    btn.addEventListener('click',()=>{
      const key=btn.getAttribute('data-clear');
      const url=new URL(window.location.href);
      if(key) url.searchParams.delete(key);
      window.location.href=url.pathname+(url.searchParams.toString()?"?"+url.searchParams.toString():"");
    });
  });

  const loadBtn=document.getElementById('loadMore');
  const cards=document.querySelector('.cards');
  const statusEl=document.getElementById('loadMoreStatus');
  function announce(msg){if(statusEl){statusEl.textContent=msg;}}
  if(loadBtn && cards){
    loadBtn.addEventListener('click',async()=>{
      const next=loadBtn.getAttribute('data-next');
      if(!next){loadBtn.disabled=true;return}
      const oldText=loadBtn.textContent; loadBtn.disabled=true; loadBtn.textContent='Cargando...'; announce('Cargando más productos');
      try{const res=await fetch(next,{credentials:'same-origin'}); if(!res.ok) throw new Error('HTTP '+res.status); const html=await res.text(); const parser=new DOMParser(); const doc=parser.parseFromString(html,'text/html'); const newCards=doc.querySelectorAll('.cards > .card'); newCards.forEach(n=>cards.appendChild(n)); const newBtn=doc.getElementById('loadMore'); const newNext=newBtn?newBtn.getAttribute('data-next'):''; if(newNext){loadBtn.setAttribute('data-next',newNext);loadBtn.disabled=false;loadBtn.textContent=oldText;announce('Se cargaron más productos');} else {loadBtn.setAttribute('data-next','');loadBtn.disabled=true;loadBtn.textContent='No hay más';announce('No hay más productos');}}
      catch(err){loadBtn.disabled=false; loadBtn.textContent=oldText; announce('Error al cargar');}
    });
  }
})();

// Registrar Service Worker en idle para no bloquear carga
if ('serviceWorker' in navigator) {
  const registerSW = () => navigator.serviceWorker.register('/public/sw.js').catch(()=>{});
  if (window.requestIdleCallback) requestIdleCallback(registerSW, {timeout: 2000});
  else window.addEventListener('load', registerSW, {once:true});
}


