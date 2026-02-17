// UI scripts unificados (module)
// - Nav drawer
// - Carousel home
// - Modal "Cómo comprar"
// - Products drawer/sheet + load more
// - Carrito: cálculo de envío
// - SW registration (idle)

// Sistema de notificaciones Toast
function showToast(message, type = 'info', duration = 3500) {
  const toast = document.createElement('div');
  toast.className = `toast toast-${type}`;
  toast.style.cssText = `position:fixed;top:20px;right:20px;z-index:99999;padding:16px 24px;border-radius:12px;font-weight:600;font-size:14px;box-shadow:0 8px 24px rgba(0,0,0,0.4);display:flex;align-items:center;gap:12px;min-width:300px;max-width:500px;animation:slideInRight 0.3s ease-out;backdrop-filter:blur(8px)`;
  
  const colors = {
    success: 'background:linear-gradient(135deg,#10b981,#059669);color:white',
    error: 'background:linear-gradient(135deg,#ef4444,#dc2626);color:white',
    warning: 'background:linear-gradient(135deg,#f59e0b,#d97706);color:white',
    info: 'background:linear-gradient(135deg,#6366f1,#8b5cf6);color:white'
  };
  
  const icons = {
    success: '✓',
    error: '✕',
    warning: '⚠',
    info: 'ℹ'
  };
  
  toast.style.cssText += colors[type] || colors.info;
  toast.innerHTML = `<span style="font-size:20px">${icons[type] || icons.info}</span><span>${message}</span>`;
  
  document.body.appendChild(toast);
  
  setTimeout(() => {
    toast.style.animation = 'slideOutRight 0.3s ease-out';
    setTimeout(() => toast.remove(), 300);
  }, duration);
}

// Función helper para formatear precios con puntos de miles
function formatPrice(num) {
  const str = num.toFixed(2);
  const parts = str.split('.');
  const intStr = parts[0];
  const decStr = parts[1];
  
  // Agregar puntos de miles a la parte entera
  let result = '';
  for (let i = 0; i < intStr.length; i++) {
    if (i > 0 && (intStr.length - i) % 3 === 0) {
      result += '.';
    }
    result += intStr[i];
  }
  
  // Si los decimales son "00", no mostrarlos
  if (decStr === '00') {
    return result;
  }
  return result + '.' + decStr;
}

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
  // Buscar slides dentro de links o directamente
  const links=[...document.querySelectorAll('#heroCarousel .hc-slide-link')];
  const directSlides=[...document.querySelectorAll('#heroCarousel .hc-slide:not(.hc-slide-link .hc-slide)')];
  const slides = links.map(l => l.querySelector('.hc-slide')).filter(s => s != null).concat(directSlides);
  const dotsC=document.getElementById('hcDots');
  if(slides.length && dotsC){
    slides.forEach((_,i)=>{const d=document.createElement('div');d.className='dot'+(i===0?' active':'');d.dataset.i=i;d.onclick=()=>go(i,true);dotsC.appendChild(d);});
    let idx=0,timer=null;
    function go(n,manual){
      // Quitar activo de slide, link y dot actual
      if(slides[idx].closest('.hc-slide-link')){
        slides[idx].closest('.hc-slide-link').classList.remove('active');
      }
      slides[idx].classList.remove('active');
      dotsC.children[idx].classList.remove('active');
      // Calcular nuevo índice
      idx=(n+slides.length)%slides.length;
      // Agregar activo a slide, link y dot nuevo
      if(slides[idx].closest('.hc-slide-link')){
        slides[idx].closest('.hc-slide-link').classList.add('active');
      }
      slides[idx].classList.add('active');
      dotsC.children[idx].classList.add('active');
      if(manual){restart();}
    }
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
  document.addEventListener('keydown',e=>{if(e.key==='Escape' && bd && !bd.hidden) close();});
})();

// Products sheet + chips + load more
(function(){
  const filterBtn=document.querySelector('.btn-filter');
  const sortBtn=document.querySelector('.btn-sort');
  const filterSheet=document.getElementById('filterSheet');
  const filterSheetClose=filterSheet?filterSheet.querySelector('.sheet-close'):null;
  const sortSheet=document.getElementById('sortSheet');
  const sortSheetClose=sortSheet?sortSheet.querySelector('.sheet-close'):null;
  const form=document.getElementById('filtersForm');
  const sortInput=document.getElementById('sortInput');

  function openFilterSheet(){
    if(!filterSheet) return;
    filterSheet.hidden=false;
    filterBtn&&filterBtn.setAttribute('aria-expanded','true');
  }
  function closeFilterSheet(){
    if(!filterSheet) return;
    filterSheet.hidden=true;
    filterBtn&&filterBtn.setAttribute('aria-expanded','false');
  }
  function openSortSheet(){
    if(!sortSheet) return;
    sortSheet.hidden=false;
    sortBtn&&sortBtn.setAttribute('aria-expanded','true');
  }
  function closeSortSheet(){
    if(!sortSheet) return;
    sortSheet.hidden=true;
    sortBtn&&sortBtn.setAttribute('aria-expanded','false');
  }

  filterBtn&&filterBtn.addEventListener('click',()=>{
    if(filterSheet&&filterSheet.hidden){
      openFilterSheet();
    }else{
      closeFilterSheet();
    }
  });
  filterSheetClose&&filterSheetClose.addEventListener('click',closeFilterSheet);
  sortBtn&&sortBtn.addEventListener('click',()=>{
    if(sortSheet&&sortSheet.hidden){
      openSortSheet();
    }else{
      closeSortSheet();
    }
  });
  sortSheetClose&&sortSheetClose.addEventListener('click',closeSortSheet);
  window.addEventListener('keydown',(e)=>{
    if(e.key==='Escape'){
      if(filterSheet&&!filterSheet.hidden) closeFilterSheet();
      if(sortSheet&&!sortSheet.hidden) closeSortSheet();
    }
  });

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

  const cards=document.querySelector('.cards');
  const statusEl=document.getElementById('loadMoreStatus');
  const infiniteScrollTrigger=document.getElementById('infiniteScrollTrigger');
  const loadingIndicator=document.getElementById('loadingIndicator');
  const endMessage=document.getElementById('endMessage');
  
  // Validación/rehuso de imágenes rotas en tarjetas de productos
  function chooseNextValidImage(imgEl){
    if(!imgEl) return false;
    const srcset=imgEl.getAttribute('srcset')||'';
    const urls=[imgEl.getAttribute('src')||''];
    // intentar extraer URLs del srcset
    srcset.split(',').forEach(part=>{ const url=(part||'').trim().split(' ')[0]; if(url) urls.push(url); });
    let idx=0;
    function tryNext(){
      idx++;
      // no hay más: remover la card
      if(idx>=urls.length){ const card=imgEl.closest('.card'); if(card) card.remove(); return; }
      imgEl.src=urls[idx];
    }
    imgEl.addEventListener('error', tryNext, {once:false});
    return true;
  }
  // Aplicar a imágenes existentes
  document.querySelectorAll('.cards .card-img').forEach(img=>chooseNextValidImage(img));
  function announce(msg){if(statusEl){statusEl.textContent=msg;}}
  
  // Scroll infinito con Intersection Observer
  if(infiniteScrollTrigger && cards){
    let isLoading=false;
    let hasMore=true;
    
    async function loadMoreProducts(){
      if(isLoading || !hasMore) return;
      
      const next=infiniteScrollTrigger.getAttribute('data-next');
      if(!next){
        hasMore=false;
        if(endMessage) endMessage.style.display='block';
        return;
      }
      
      isLoading=true;
      if(loadingIndicator) loadingIndicator.style.display='block';
      announce('Cargando más productos');
      
      try{
        const res=await fetch(next,{credentials:'same-origin'});
        if(!res.ok) throw new Error('HTTP '+res.status);
        const html=await res.text();
        const parser=new DOMParser();
        const doc=parser.parseFromString(html,'text/html');
        const newCards=doc.querySelectorAll('.cards > .card');
        
        newCards.forEach(n=>{
          cards.appendChild(n);
          const img=n.querySelector('.card-img');
          if(img) chooseNextValidImage(img);
        });
        
        const newTrigger=doc.getElementById('infiniteScrollTrigger');
        const newNext=newTrigger?newTrigger.getAttribute('data-next'):'';
        
        const resultCount=document.querySelector('.result-count');
        if(resultCount){
          const currentCards=document.querySelectorAll('.cards > .card').length;
          const totalMatch=resultCount.textContent.match(/de (\d+)/);
          const total=totalMatch?totalMatch[1]:currentCards;
          resultCount.textContent=currentCards+' resultados de '+total;
        }
        
        if(newNext){
          infiniteScrollTrigger.setAttribute('data-next',newNext);
          announce('Se cargaron más productos');
        } else {
          hasMore=false;
          infiniteScrollTrigger.setAttribute('data-next','');
          if(endMessage) endMessage.style.display='block';
          announce('No hay más productos para cargar');
        }
      } catch(err){
        announce('Error al cargar');
        console.error('Error loading products:', err);
      } finally {
        isLoading=false;
        if(loadingIndicator) loadingIndicator.style.display='none';
      }
    }
    
    // Usar Intersection Observer para detectar cuando el usuario llega al trigger
    const observer=new IntersectionObserver((entries)=>{
      entries.forEach(entry=>{
        if(entry.isIntersecting && hasMore && !isLoading){
          loadMoreProducts();
        }
      });
    }, {
      rootMargin:'400px' // Cargar cuando está a 400px de llegar al trigger
    });
    
    observer.observe(infiniteScrollTrigger);
  }
})();

// Carrito: cálculo de envío y total (compatible con CSP)
(function(){
  const form=document.getElementById('checkoutForm'); if(!form) return;
  const shipRadios=form.querySelectorAll('input[name="shipping"]');
  const paymentRadios=form.querySelectorAll('input[name="payment_method"]');
  const envioGroup=document.getElementById('envioGroup');
  const cadeteGroup=document.getElementById('cadeteGroup');
  const provinceSelect=document.getElementById('provinceSelect');
  const shipCostEl=document.getElementById('shipCost');
  const subtotalEl=document.getElementById('subtotalVal');
  const discountEl=document.getElementById('discount');
  const discountRow=document.getElementById('discountRow');
  const finalTotalEl=document.getElementById('finalTotal');
  if(!subtotalEl || !finalTotalEl || !shipCostEl) return;
  const base=parseFloat((subtotalEl.textContent || '').replace(/[^0-9.,]/g,'').replace(',','.'))||0;
  const CADETE_COST=5000;
  const COSTS=(()=>{ const m={}; document.querySelectorAll('#pcData [data-prov]').forEach(n=>{ const k=n.getAttribute('data-prov'); const v=parseFloat(n.getAttribute('data-cost')||'0'); if(k){ m[k]=v; } }); return m; })();
  const phone = form.querySelector('input[name="phone"]');
  const addrCadete = form.querySelector('input[name="address_cadete"]');
  const addrEnvio = form.querySelector('input[name="address_envio"]');
  const postal = form.querySelector('input[name="postal_code"]');
  const dni = form.querySelector('input[name="dni"]');
  function setRequired(el,flag){ if(!el) return; if(flag){el.setAttribute('required','required')} else {el.removeAttribute('required')} }
  function updateRadioBorder(elements){
    elements.forEach(el=>{
      if(el.checked) el.closest('label').style.borderColor='#6366f1';
      else el.closest('label').style.borderColor='var(--border)';
    });
  }
  function calcCost(){
    let method='retiro'; shipRadios.forEach(r=>{ if(r.checked) method=r.value });
    let paymentMethod='efectivo'; paymentRadios.forEach(r=>{ if(r.checked) paymentMethod=r.value });
    updateRadioBorder(shipRadios);
    updateRadioBorder(paymentRadios);
    if(envioGroup) envioGroup.style.display='none'; if(cadeteGroup) cadeteGroup.style.display='none';
    setRequired(phone,false); setRequired(addrCadete,false); setRequired(addrEnvio,false); setRequired(provinceSelect,false); setRequired(postal,false); setRequired(dni,false);
    let cost=0;
    if(method==='envio'){
      if(envioGroup) envioGroup.style.display='flex';
      setRequired(phone,true); setRequired(addrEnvio,true); setRequired(provinceSelect,true); setRequired(postal,true); setRequired(dni,true);
      const prov=provinceSelect?provinceSelect.value:''; if(prov && COSTS[prov]!=null){cost=COSTS[prov];}
    } else if(method==='cadete') {
      if(cadeteGroup) cadeteGroup.style.display='flex';
      setRequired(phone,true); setRequired(addrCadete,true);
      cost=CADETE_COST;
    }
    if(shipCostEl) shipCostEl.textContent='$'+formatPrice(cost);
    const withShip=(base+cost);
    const discount=(paymentMethod==='transferencia'? withShip*0.1 : 0);
    if(discountRow){
      if(discount>0){
        discountRow.style.display='flex';
        if(discountEl) discountEl.textContent='-$'+formatPrice(discount);
      } else {
        discountRow.style.display='none';
        if(discountEl) discountEl.textContent='$0';
      }
    }
    const final=(withShip-discount);
    if(finalTotalEl) finalTotalEl.textContent='$'+formatPrice(final);
  }
  shipRadios.forEach(r=>r.addEventListener('change',calcCost));
  paymentRadios.forEach(r=>r.addEventListener('change',calcCost));
  provinceSelect && provinceSelect.addEventListener('change',calcCost);
  calcCost();
  
  // Popup de confirmación de compra
  const modal=document.getElementById('confirmPurchaseModal');
  const confirmTotal=document.getElementById('confirmTotal');
  const confirmProductCount=document.getElementById('confirmProductCount');
  const confirmBtn=document.getElementById('confirmPurchaseBtn');
  const cancelBtn=document.getElementById('confirmCancelBtn');
  if(modal && confirmTotal && confirmProductCount && confirmBtn && cancelBtn){
    let shouldSubmit=false;
    function countProducts(){
      const rows=document.querySelectorAll('.cart-table tbody tr');
      let totalQty=0;
      rows.forEach(row=>{
        const qtySpan=row.querySelector('span[style*="min-width:36px"]');
        if(qtySpan){
          const qty=parseInt(qtySpan.textContent.trim())||0;
          totalQty+=qty;
        }
      });
      return totalQty;
    }
    function getTotal(){
      if(!finalTotalEl) return '$0.00';
      return finalTotalEl.textContent.trim();
    }
    function showModal(){
      const total=getTotal();
      const productCount=countProducts();
      confirmTotal.textContent=total;
      confirmProductCount.textContent=productCount+(productCount===1?' producto':' productos');
      modal.style.display='flex';
      document.body.style.overflow='hidden';
    }
    function hideModal(){
      modal.style.display='none';
      document.body.style.overflow='';
    }
    form.addEventListener('submit',function(e){
      if(!shouldSubmit){
        e.preventDefault();
        showModal();
      }
    });
    confirmBtn.addEventListener('click',function(){
      shouldSubmit=true;
      hideModal();
      form.submit();
    });
    cancelBtn.addEventListener('click',hideModal);
    modal.addEventListener('click',function(e){
      if(e.target===modal) hideModal();
    });
    document.addEventListener('keydown',function(e){
      if(e.key==='Escape' && modal.style.display==='flex') hideModal();
    });
  }
})();

// Registrar Service Worker en idle para no bloquear carga
if ('serviceWorker' in navigator) {
  const registerSW = () => navigator.serviceWorker.register('/public/sw.js').catch(()=>{});
  if (window.requestIdleCallback) requestIdleCallback(registerSW, {timeout: 2000});
  else window.addEventListener('load', registerSW, {once:true});
}

// Producto: carrusel de imágenes (CSP-safe, sin inline)
(function(){
  const root=document.getElementById('pdCarousel'); if(!root) return;
  const slides=[...root.querySelectorAll('.pd-slide')];
  const thumbs=[...root.querySelectorAll('.pd-thumb')];
  const total=slides.length; if(total===0) return;
  const prevBtn=root.querySelector('.pd-nav.prev');
  const nextBtn=root.querySelector('.pd-nav.next');
  if(total===1){ if(prevBtn) prevBtn.style.display='none'; if(nextBtn) nextBtn.style.display='none'; const tb=root.querySelector('.pd-thumbs'); if(tb) tb.style.display='none'; return; }
  let idx=0; let auto=null; const AUTOPLAY=5000;
  function show(n){
    slides[idx].classList.remove('active'); thumbs[idx].classList.remove('active');
    idx=(n+total)%total;
    slides[idx].classList.add('active'); thumbs[idx].classList.add('active');
  }
  function next(){ show(idx+1); }
  function prev(){ show(idx-1); }
  function start(){ auto=setInterval(next,AUTOPLAY); }
  function restart(){ if(auto) clearInterval(auto); start(); }
  root.addEventListener('click', (e)=>{
    const t=e.target.closest('.pd-thumb');
    if(t){ const i=parseInt(t.getAttribute('data-index')||'-1'); if(!isNaN(i)) { show(i); restart(); } return; }
    if(e.target.closest('.pd-nav.prev')){ prev(); restart(); return; }
    if(e.target.closest('.pd-nav.next')){ next(); restart(); return; }
  });
  // Swipe táctil
  let startX=0,isSwiping=false;
  root.addEventListener('touchstart',e=>{startX=e.touches[0].clientX;isSwiping=true;},{passive:true});
  root.addEventListener('touchmove',e=>{ if(!isSwiping) return; const dx=e.touches[0].clientX-startX; if(Math.abs(dx)>10){ e.preventDefault(); } },{passive:false});
  root.addEventListener('touchend',e=>{ if(!isSwiping) return; isSwiping=false; const dx=e.changedTouches[0].clientX-startX; if(Math.abs(dx)>50){ if(dx>0) prev(); else next(); restart(); } },{passive:true});
  // Teclado
  root.addEventListener('keydown',e=>{ if(e.key==='ArrowRight'){ next(); restart(); } else if(e.key==='ArrowLeft'){ prev(); restart(); } });
  root.tabIndex=0; start();
})();

// Producto: compartir (CSP-safe)
(function(){
  const bar=document.getElementById('pdShareBar'); if(!bar) return;
  const msgEl=document.getElementById('shareMsg');
  const titleEl=document.querySelector('.pd-title');
  const productName=titleEl?((titleEl.textContent||'').trim()):'';
  const buildURL=()=>window.location.href.split('#')[0];
  function flash(t){ if(!msgEl) return; msgEl.textContent=t; msgEl.style.display='inline'; msgEl.classList.add('show'); setTimeout(()=>{ msgEl.style.display='none'; msgEl.classList.remove('show'); },2500); }
  bar.addEventListener('click', async e=>{
    const btn=e.target.closest('.share-btn'); if(!btn) return;
    e.preventDefault();
    const url=buildURL();
    const text=`Mira este producto: ${productName} - ${url}`;
    const kind=btn.getAttribute('data-share');
    if(kind==='copy'){
      let copied=false;
      try{ if(navigator.clipboard && window.isSecureContext){ await navigator.clipboard.writeText(text); copied=true; } }catch{}
      if(!copied){ try{ const ta=document.createElement('textarea'); ta.value=text; ta.style.position='fixed'; ta.style.top='-1000px'; document.body.appendChild(ta); ta.focus(); ta.select(); document.execCommand('copy'); ta.remove(); copied=true; }catch{} }
      flash(copied?'Enlace copiado':'No se pudo copiar');
    } else if(kind==='whatsapp'){
      const enc=encodeURIComponent(text);
      const w1=`https://wa.me/?text=${enc}`;
      const w2=`https://api.whatsapp.com/send?text=${enc}`;
      const win=window.open(w1,'_blank','noopener');
      if(!win || win.closed){ setTimeout(()=>{ window.open(w2,'_blank','noopener') || (location.href=w1); },50); }
    } else if(kind==='instagram'){
      const shareData={title:productName,text:`Mira este producto: ${productName}`,url};
      if(navigator.share){ try{ await navigator.share(shareData); }catch{} }
      else {
        let copied=false; try{ if(navigator.clipboard && window.isSecureContext){ await navigator.clipboard.writeText(text); copied=true; } }catch{}
        if(!copied){ try{ const ta=document.createElement('textarea'); ta.value=text; ta.style.position='fixed'; ta.style.top='-1000px'; document.body.appendChild(ta); ta.focus(); ta.select(); document.execCommand('copy'); ta.remove(); copied=true; }catch{} }
        window.open('https://www.instagram.com/direct/new/','_blank','noopener');
        flash(copied?'Texto copiado. Pega en Instagram':'No se pudo copiar');
      }
    } else if(kind==='x'){
      const u=encodeURIComponent(url);
      const t=encodeURIComponent(`Mira este producto: ${productName}`);
      const xUrl=`https://twitter.com/intent/tweet?url=${u}&text=${t}`;
      window.open(xUrl,'_blank','noopener');
    } else if(kind==='threads'){
      const u=encodeURIComponent(url);
      const t=encodeURIComponent(`Mira este producto: ${productName}`);
      const thUrl=`https://www.threads.net/intent/post?text=${t}%20${u}`;
      window.open(thUrl,'_blank','noopener');
    }
  });
})();

// Producto: selector de color (CSP-safe, sin inline)
(function(){
  const form=document.querySelector('.pd-form'); if(!form) return;
  const inp=document.getElementById('colorInput'); if(!inp) return;
  const swc=document.getElementById('cpSwatches');
  const preview=document.getElementById('colorPreview');
  const customRow=document.getElementById('customColorRow');
  const customInput=document.getElementById('customColorInput');
  const otherBtn=form.querySelector('.swatch-other');
  const defaultColor=inp.value||'#111827';

  function setPreview(c){ if(preview){ preview.style.background = (c&&c.trim()) || defaultColor; } }
  function setActive(btn){ if(!swc) return; swc.querySelectorAll('.swatch').forEach(x=>{ x.classList.remove('active'); x.setAttribute('aria-checked','false'); }); if(btn){ btn.classList.add('active'); btn.setAttribute('aria-checked','true'); } }

  if(swc){
    swc.addEventListener('click',e=>{
      const b=e.target.closest('.swatch'); if(!b) return;
      e.preventDefault();
      setActive(b);
      if(b.classList.contains('swatch-other')){
        if(customRow) customRow.style.display='';
        if(customInput) customInput.focus();
        inp.value = (customInput && customInput.value.trim()) || '';
      } else {
        if(customRow) customRow.style.display='none';
        inp.value = b.getAttribute('data-color') || '';
      }
      setPreview(inp.value);
    });
  }

  if(customInput){
    customInput.addEventListener('input',()=>{
      const val=customInput.value.trim();
      inp.value=val;
      if(otherBtn && val){ otherBtn.style.setProperty('--c', val); }
      setPreview(val);
    });
  }

  form.addEventListener('submit',()=>{ if(!inp.value.trim()){ inp.value=defaultColor; } });
  setPreview(defaultColor);
})();

// Producto: sticky mobile CTA
(function(){
  const stickyBar=document.getElementById('pdMobileSticky');
  const formSection=document.querySelector('.pd-form');
  if(!stickyBar || !formSection) return;
  
  let lastScroll=0;
  const observerOptions={root:null,threshold:0,rootMargin:'-100px 0px 0px 0px'};
  
  const observer=new IntersectionObserver(entries=>{
    entries.forEach(entry=>{
      const currentScroll=window.pageYOffset;
      const scrollingDown=currentScroll>lastScroll;
      if(!entry.isIntersecting && scrollingDown){
        stickyBar.style.display='block';
      }else{
        stickyBar.style.display='none';
      }
      lastScroll=currentScroll;
    });
  },observerOptions);
  
  if(window.innerWidth<768){
    observer.observe(formSection);
  }
  
  window.addEventListener('resize',()=>{
    if(window.innerWidth>=768){
      stickyBar.style.display='none';
      observer.disconnect();
    }else{
      observer.observe(formSection);
    }
  });
})();

// Admin: productos (form + tabla + dropzone) sin inline JS
(function(){
  const form=document.getElementById('prodForm'); if(!form) return;
  const tbl=document.getElementById('prodTable');
  const searchInput=document.getElementById('prodSearch');
  const token=(form.getAttribute('data-token')||'').trim();
  const fSlug=document.getElementById('pfSlug');
  const fName=document.getElementById('pfName');
  const fCat=document.getElementById('pfCategory');
  const fDesc=document.getElementById('pfShort');
  const fPrice=document.getElementById('pfPrice');
  const fReady=document.getElementById('pfReady');
  const fWidth=document.getElementById('pfWidth');
  const fHeight=document.getElementById('pfHeight');
  const fDepth=document.getElementById('pfDepth');
  const fObservation=document.getElementById('pfObservation');
  const fGrams=document.getElementById('pfGrams');
  const fHours=document.getElementById('pfHours');
  const fGrossPrice=document.getElementById('pfGrossPrice');
  const fProfit=document.getElementById('pfProfit');
  const btnDel=document.getElementById('pfDelete');
  const btnReset=document.getElementById('pfReset');
  const btnSubmit=document.getElementById('pfSubmit');
  const modeBadge=document.getElementById('formMode');
  const imagesInput=document.getElementById('pfImages');
  const dropZone=document.getElementById('dropZone');
  const preview=document.getElementById('preview');
  const dzCount=document.querySelector('.dz-count');
  const btnManage=document.getElementById('pfManageImages');
  const btnManageAlt=document.getElementById('pfManageImagesAlt');
  const btnDelSelected=document.getElementById('btnDelSelected');
  const selCount=document.getElementById('selCount');
  const selectedIDs=new Set();
  const topbar=document.querySelector('.topbar');
  const btnRepair=document.getElementById('btnRepairImages');
  const repStatus=document.getElementById('repStatus');

  function getCurrentSlug(){ return (fSlug && fSlug.value.trim()) || ''; }
  function updateImgsCountForSlug(slug, count){
    if(!slug) return;
    const tr=document.querySelector('tr[data-slug="'+CSS.escape(slug)+'"]');
    if(!tr) return;
    const tds=tr.querySelectorAll('td');
    if(tds && tds[12]){ tds[12].textContent = String(count|0); }
  }
  async function reloadProductAndSync(){
    const slug=getCurrentSlug(); if(!slug) return;
    try{
      const res=await fetch('/api/products/'+encodeURIComponent(slug),{headers: token? {Authorization:'Bearer '+token}:{}});
      if(!res.ok) return;
      const p=await res.json();
      const imgs=(p && Array.isArray(p.Images))? p.Images: [];
      renderGallery(imgs);
      updateImgsCountForSlug(slug, imgs.length|0);
    }catch{}
  }
  async function syncImgsCountFromFormGallery(){
    const slug=getCurrentSlug(); if(!slug) return;
    // Preferir el valor real desde API (ya filtrado por el servidor)
    try{
      const res=await fetch('/api/products/'+encodeURIComponent(slug),{headers: token? {Authorization:'Bearer '+token}:{}});
      if(res.ok){ const p=await res.json(); const count=(p && Array.isArray(p.Images))? p.Images.length : 0; updateImgsCountForSlug(slug, count); return; }
    }catch{}
    // Fallback al conteo del DOM
    const count=document.querySelectorAll('#imgGallery .img-card').length; updateImgsCountForSlug(slug, count);
  }

  function scrollToForm(){
    const offset=(topbar && topbar.getBoundingClientRect().height)||72;
    const y=form.getBoundingClientRect().top + window.pageYOffset - offset - 10;
    window.scrollTo({ top: y>0? y: 0, behavior: 'smooth' });
    if(fName && typeof fName.focus==='function'){
      try{ fName.focus({preventScroll:true}); }catch(_e){}
    }
  }

  function setModeEdit(edit){
    if(!modeBadge||!btnSubmit) return;
    if(edit){ modeBadge.textContent='Edición'; modeBadge.classList.remove('create'); modeBadge.classList.add('edit'); btnSubmit.textContent='Actualizar'; if(btnManage) btnManage.style.display=''; }
    else { modeBadge.textContent='Creación'; modeBadge.classList.remove('edit'); modeBadge.classList.add('create'); btnSubmit.textContent='Crear'; if(btnManage) btnManage.style.display='none'; }
  }
  function fill(p){
    if(!p) return;
    if(fSlug) fSlug.value=p.Slug||''; if(fName) fName.value=p.Name||''; if(fCat) fCat.value=p.Category||''; if(fDesc) fDesc.value=p.ShortDesc||''; if(fPrice) fPrice.value=p.BasePrice!=null?p.BasePrice:''; if(fReady) fReady.checked=!!p.ReadyToShip; if(fWidth) fWidth.value=p.WidthMM||0; if(fHeight) fHeight.value=p.HeightMM||0; if(fDepth) fDepth.value=p.DepthMM||0; if(fObservation) fObservation.value=p.Observation||''; if(fGrams) fGrams.value=p.Grams||0; if(fHours) fHours.value=p.Hours||0; if(fGrossPrice) fGrossPrice.value=p.GrossPrice||0; if(fProfit) fProfit.value=p.Profit||0; if(btnDel) btnDel.style.display=''; setModeEdit(true);
    renderGallery((p && p.Images) || []);
    scrollToForm();
  }
  function clear(){ if(form) form.reset(); if(fSlug) fSlug.value=''; if(btnDel) btnDel.style.display='none'; if(btnManage) btnManage.style.display='none'; if(fWidth) fWidth.value=''; if(fHeight) fHeight.value=''; if(fDepth) fDepth.value=''; if(fObservation) fObservation.value=''; if(fGrams) fGrams.value=''; if(fHours) fHours.value=''; if(fGrossPrice) fGrossPrice.value=''; if(fProfit) fProfit.value=''; if(imagesInput) imagesInput.value=''; if(preview) preview.innerHTML=''; if(dzCount) dzCount.textContent='0 archivos'; setModeEdit(false); selectedIDs.clear(); updateSelectionUI(); renderGallery([]); }

  function renderGallery(imgs){
    const g=document.getElementById('imgGallery'); if(!g) return;
    g.innerHTML=''; selectedIDs.clear(); updateSelectionUI();
    (imgs||[]).forEach(im=>{
      if(!im || !im.URL || !im.ID) return;
      const card=document.createElement('div'); card.className='img-card'; card.dataset.id=im.ID;
      card.style.position='relative'; card.style.width='80px'; card.style.height='80px'; card.style.cursor='pointer';
      const img=document.createElement('img'); img.src=im.URL; img.alt=im.Alt||''; img.loading='lazy'; img.style.width='100%'; img.style.height='100%'; img.style.objectFit='cover'; img.style.borderRadius='10px'; img.style.border='1px solid #223140'; img.onerror=()=>{ card.remove(); };
      const del=document.createElement('button'); del.type='button'; del.textContent='✖'; del.title='Eliminar'; del.className='icon-btn danger'; del.style.position='absolute'; del.style.top='-6px'; del.style.right='-6px'; del.style.background='#ef4444'; del.style.color='#fff'; del.style.borderRadius='50%'; del.style.width='22px'; del.style.height='22px'; del.style.display='grid'; del.style.placeItems='center'; del.style.fontSize='12px'; del.style.cursor='pointer';
      const sel=document.createElement('div'); sel.textContent='✓'; sel.setAttribute('aria-hidden','true'); sel.style.position='absolute'; sel.style.left='-6px'; sel.style.top='-6px'; sel.style.width='20px'; sel.style.height='20px'; sel.style.borderRadius='50%'; sel.style.display='grid'; sel.style.placeItems='center'; sel.style.fontSize='12px'; sel.style.background='#10b981'; sel.style.color='#0b1520'; sel.style.border='1px solid #0f2b3d'; sel.style.boxShadow='0 0 0 2px #0b1520'; sel.style.opacity='0'; sel.style.transition='opacity .15s ease';
      del.addEventListener('click', async (e)=>{
        e.preventDefault();
        if(!token){ alert('Sesión no válida'); return; }
        if(!confirm('Eliminar esta imagen?')) return;
        const res=await fetch('/api/product_images/'+encodeURIComponent(im.ID), {method:'DELETE', headers:{Authorization:'Bearer '+token}});
        if(res.ok){ selectedIDs.delete(im.ID); updateSelectionUI(); card.remove(); const modalCard=document.querySelector('#imgMgrGrid .img-card[data-id="'+CSS.escape(im.ID)+'"]'); if(modalCard) modalCard.remove(); await reloadProductAndSync(); } else { alert('Error eliminando imagen'); }
      });
      del.addEventListener('click', e=>e.stopPropagation());
      card.appendChild(img); card.appendChild(del); card.appendChild(sel); g.appendChild(card);
    });
    // Sincronizar contador al render
    syncImgsCountFromFormGallery();
    g.onclick=(e)=>{
      const card=e.target.closest('.img-card'); if(!card) return;
      if(e.target.closest('button')) return;
      const id=card.dataset.id; if(!id) return;
      if(selectedIDs.has(id)){ selectedIDs.delete(id); card.style.outline=''; const b=card.querySelector('div[aria-hidden]'); if(b) b.style.opacity='0'; }
      else { selectedIDs.add(id); card.style.outline='2px solid #10b981'; const b=card.querySelector('div[aria-hidden]'); if(b) b.style.opacity='1'; }
      updateSelectionUI();
    };
  }

  function updateSelectionUI(){
    const n=selectedIDs.size; if(selCount) selCount.textContent = n>0? (n===1? '1 imagen seleccionada': n+' imágenes seleccionadas') : '';
    if(btnDelSelected) btnDelSelected.style.display = n>0? '': 'none';
  }

  if(tbl){ tbl.addEventListener('click', async e=>{
    const btn=e.target.closest('button'); if(!btn) return; const tr=btn.closest('tr'); if(!tr) return; const slug=tr.getAttribute('data-slug')||'';
    if(btn.getAttribute('data-act')==='images'){
      if(fSlug) fSlug.value=slug; openMgr(); /* evitar doble carga: solo abrir, mgrOpen llama a mgrLoad */ return;
    }
    if(btn.getAttribute('data-act')==='edit'){
      const res=await fetch('/api/products/'+encodeURIComponent(slug),{headers: token? {Authorization:'Bearer '+token}:{}}); if(res.ok){ const p=await res.json(); fill(p);} return;
    }
    if(btn.getAttribute('data-act')==='del'){
      if(!confirm('Eliminar producto y todos sus datos?')) return; const res=await fetch('/api/products/'+encodeURIComponent(slug),{method:'DELETE', headers: token? {Authorization:'Bearer '+token}:{}}); if(res.ok){ location.reload(); } else { alert('Error eliminando'); } return;
    }
  }); }

  // Filtro local por nombre, slug y categoría (case-insensitive mejorado)
  if(searchInput && tbl){
    const tbody=tbl.tBodies && tbl.tBodies[0];
    const categoryFilter=document.getElementById('categoryFilter');
    const rows=()=>tbody? Array.from(tbody.rows): [];
    
    function applyFilter(){
      const q=(searchInput.value||'').trim().toLowerCase();
      const selCat=(categoryFilter && categoryFilter.value||'').trim().toLowerCase(); // ← CASE-INSENSITIVE
      
      let visibleCount = 0;
      rows().forEach(tr=>{
        const name=(tr.cells[0] && tr.cells[0].textContent||'').toLowerCase();
        const slug=(tr.cells[1] && tr.cells[1].textContent||'').toLowerCase();
        const cat=(tr.getAttribute('data-category')||'').trim().toLowerCase(); // ← CASE-INSENSITIVE
        const imgsCount=parseInt(tr.getAttribute('data-imgs-count')||'0', 10);
        
        const matchText = !q || name.includes(q) || slug.includes(q);
        let matchCat = true;
        
        // Filtro especial para "sin imagen"
        if(selCat === 'sin imagen'){
          matchCat = imgsCount === 0;
        } else if(selCat === 'sin categoría'){
          matchCat = !cat;
        } else if(selCat){
          matchCat = cat === selCat;
        }
        
        const isVisible = matchText && matchCat;
        tr.style.display = isVisible ? '' : 'none';
        if(isVisible) visibleCount++;
      });
      
      // Actualizar contador
      updateFilterCount(visibleCount, rows().length);
    }
    
    function updateFilterCount(visible, total){
      let counterEl = document.getElementById('productsFilterCounter');
      if(!counterEl){
        counterEl = document.createElement('span');
        counterEl.id = 'productsFilterCounter';
        counterEl.style.cssText = 'font-size:13px;color:var(--muted);margin-left:8px';
        const h2 = document.querySelector('.admin-card + div h2 span');
        if(h2) h2.parentNode.appendChild(counterEl);
      }
      
      const hasFilters = searchInput.value.trim() !== '' || (categoryFilter && categoryFilter.value !== '');
      const btnClear = document.getElementById('btnClearFilters');
      
      if(hasFilters){
        counterEl.textContent = `(${visible} de ${total})`;
        counterEl.style.color = visible === 0 ? '#ef4444' : '#10b981';
        if(btnClear) btnClear.style.display = 'inline-block';
      } else {
        counterEl.textContent = '';
        if(btnClear) btnClear.style.display = 'none';
      }
      
      // Mostrar mensaje si no hay resultados
      if(visible === 0 && hasFilters){
        showNoResultsMessage();
      } else {
        hideNoResultsMessage();
      }
    }
    
    function showNoResultsMessage(){
      let msg = document.getElementById('noResultsMsg');
      if(!msg){
        msg = document.createElement('div');
        msg.id = 'noResultsMsg';
        msg.style.cssText = 'padding:40px 20px;text-align:center;background:var(--panel);border-radius:12px;margin:20px 0;border:1px solid var(--border)';
        msg.innerHTML = '<div style="font-size:48px;margin-bottom:12px">🔍</div><div style="font-size:16px;font-weight:600;margin-bottom:8px;color:var(--text)">No se encontraron productos</div><div style="font-size:13px;color:var(--muted)">Intentá con otros términos de búsqueda o categoría</div>';
        const table = document.getElementById('prodTable');
        if(table) table.parentNode.insertBefore(msg, table);
      }
      msg.style.display = 'block';
      const table = document.getElementById('prodTable');
      if(table) table.style.display = 'none';
    }
    
    function hideNoResultsMessage(){
      const msg = document.getElementById('noResultsMsg');
      if(msg) msg.style.display = 'none';
      const table = document.getElementById('prodTable');
      if(table) table.style.display = '';
    }
    
    searchInput.addEventListener('input', applyFilter);
    if(categoryFilter) categoryFilter.addEventListener('change', applyFilter);
    
    // Botón limpiar filtros
    const btnClear = document.getElementById('btnClearFilters');
    if(btnClear){
      btnClear.addEventListener('click', function(){
        searchInput.value = '';
        if(categoryFilter) categoryFilter.value = '';
        applyFilter();
      });
    }
  }

  form.addEventListener('submit', async e=>{
    e.preventDefault();
    const submitBtn = form.querySelector('button[type="submit"]');
    if(submitBtn) {
      submitBtn.disabled = true;
      submitBtn.textContent = 'Guardando...';
    }
    
    const slug=(fSlug&&fSlug.value.trim())||'';
    const payload={ name:(fName&&fName.value.trim())||'', category:(fCat&&fCat.value.trim())||'', short_desc:(fDesc&&fDesc.value)||'', base_price:parseFloat((fPrice&&fPrice.value)||'0'), ready_to_ship:!!(fReady&&fReady.checked), width_mm:parseFloat((fWidth&&fWidth.value)||'0'), height_mm:parseFloat((fHeight&&fHeight.value)||'0'), depth_mm:parseFloat((fDepth&&fDepth.value)||'0'), observation:(fObservation&&fObservation.value)||'', grams:parseFloat((fGrams&&fGrams.value)||'0'), hours:parseFloat((fHours&&fHours.value)||'0'), gross_price:parseFloat((fGrossPrice&&fGrossPrice.value)||'0'), profit:parseFloat((fProfit&&fProfit.value)||'0') };
    
    if(!payload.name){ 
      showToast('El nombre del producto es requerido', 'error');
      if(submitBtn) { submitBtn.disabled = false; submitBtn.textContent = slug ? 'Actualizar' : 'Crear'; }
      return; 
    }
    if(payload.base_price<0){ 
      showToast('El precio debe ser mayor o igual a cero', 'error');
      if(submitBtn) { submitBtn.disabled = false; submitBtn.textContent = slug ? 'Actualizar' : 'Crear'; }
      return; 
    }
    
    let method='POST', url='/api/products'; 
    if(slug){ method='PUT'; url='/api/products/'+encodeURIComponent(slug); }
    
    try {
      const res=await fetch(url,{method, headers:Object.assign({'Content-Type':'application/json'}, token? {Authorization:'Bearer '+token}:{}) , body:JSON.stringify(payload)});
      
      if(!res.ok){ 
        const errorText = await res.text().catch(() => 'Error desconocido');
        showToast(`Error al guardar el producto: ${errorText}`, 'error', 5000);
        if(submitBtn) { submitBtn.disabled = false; submitBtn.textContent = slug ? 'Actualizar' : 'Crear'; }
        return; 
      }
      
      const prod=await res.json(); 
      const finalSlug=(prod&&prod.Slug)||slug;
      
      if(imagesInput && imagesInput.files && imagesInput.files.length>0){
        showToast('Producto guardado, subiendo imágenes...', 'info', 2000);
        const fd=new FormData(); 
        fd.append('existing_slug', finalSlug);
        for(const f of imagesInput.files){ fd.append('images', f); }
        
        const upRes=await fetch('/api/products/upload',{method:'POST', headers: token? {Authorization:'Bearer '+token}:{}, body:fd});
        
        if(!upRes.ok){ 
          showToast('Producto guardado, pero hubo un error al subir las imágenes', 'warning', 4000);
          setTimeout(() => location.reload(), 2000);
          return; 
        }
      }
      
      showToast(`Producto ${slug ? 'actualizado' : 'creado'} exitosamente!`, 'success', 2000);
      setTimeout(() => location.reload(), 1500);
      
    } catch(err) {
      showToast('Error de conexión. Verificá tu internet e intentá nuevamente.', 'error', 5000);
      if(submitBtn) { submitBtn.disabled = false; submitBtn.textContent = slug ? 'Actualizar' : 'Crear'; }
    }
  });

  if(btnReset) btnReset.addEventListener('click', clear);
  if(btnDel) btnDel.addEventListener('click', async ()=>{ const slug=(fSlug&&fSlug.value.trim())||''; if(!slug) return; if(!confirm('Eliminar producto y sus imágenes?')) return; const res=await fetch('/api/products/'+encodeURIComponent(slug),{method:'DELETE', headers: token? {Authorization:'Bearer '+token}:{}}); if(res.ok){ clear(); location.reload(); } else { alert('Error'); } });

  function refreshPreview(){ if(!preview||!imagesInput||!dzCount) return; preview.innerHTML=''; const files=Array.from(imagesInput.files||[]); dzCount.textContent=files.length+(files.length===1?' archivo':' archivos'); files.slice(0,6).forEach(f=>{ const r=new FileReader(); r.onload=ev=>{ const img=document.createElement('img'); img.src=ev.target.result; img.alt=f.name; img.style.width='52px'; img.style.height='52px'; img.style.objectFit='cover'; img.style.borderRadius='10px'; img.style.border='1px solid #223140'; preview.appendChild(img); }; r.readAsDataURL(f); }); }
  if(imagesInput) imagesInput.addEventListener('change', refreshPreview);
  if(dropZone){ ['dragenter','dragover'].forEach(ev=>dropZone.addEventListener(ev,e=>{e.preventDefault(); dropZone.classList.add('drag');})); ['dragleave','drop'].forEach(ev=>dropZone.addEventListener(ev,e=>{e.preventDefault(); dropZone.classList.remove('drag');})); dropZone.addEventListener('drop', e=>{ const files=[...e.dataTransfer.files].filter(f=>f.type.startsWith('image/')); if(files.length){ const dt=new DataTransfer(); files.forEach(f=>dt.items.add(f)); if(imagesInput) imagesInput.files=dt.files; refreshPreview(); }}); }

  // Botón: Refresh images (repara imágenes huérfanas en server y recarga la página)
  if(btnRepair){
    btnRepair.addEventListener('click', async ()=>{
      if(repStatus) repStatus.textContent='Revisando...';
      try{
        const res=await fetch('/admin/repair_images',{credentials:'same-origin'});
        const txt=await res.text();
        if(!res.ok){ throw new Error('HTTP '+res.status+': '+txt); }
        if(repStatus) repStatus.textContent='Listo. Recargando...';
        setTimeout(()=>location.reload(), 600);
      }catch(err){ if(repStatus) repStatus.textContent='Error: '+(err&&err.message||''); }
    });
  }

  // ===== Modal gestor de imágenes =====
  const overlay=document.getElementById('imgMgrOverlay');
  const grid=document.getElementById('imgMgrGrid');
  const emptyLbl=document.getElementById('imgMgrEmpty');
  const mgrSelCount=document.getElementById('imgMgrSelCount');
  const mgrDelBtn=document.getElementById('imgMgrDelSelected');
  const mgrClose=document.getElementById('imgMgrClose');
  const mgrClose2=document.getElementById('imgMgrClose2');
  const mgrSelected=new Set();
  let mgrLoadSeq=0;

  function mgrUpdateSel(){ const n=mgrSelected.size; if(mgrSelCount) mgrSelCount.textContent=n? (n===1?'1 imagen seleccionada':n+' imágenes seleccionadas') : ''; if(mgrDelBtn) mgrDelBtn.style.display=n? '':'none'; }
  function mgrOpen(){ if(!overlay) return; overlay.style.display='flex'; mgrLoad(); }
  function mgrCloseFn(){ if(!overlay) return; overlay.style.display='none'; mgrSelected.clear(); mgrUpdateSel(); }
  async function mgrLoad(){
    const slug=(fSlug && fSlug.value.trim())||''; if(!slug) return;
    const mySeq=++mgrLoadSeq;
    if(grid){ grid.innerHTML=''; }
    mgrSelected.clear(); mgrUpdateSel(); if(emptyLbl) emptyLbl.style.display='none';
    const res=await fetch('/api/products/'+encodeURIComponent(slug),{headers: token? {Authorization:'Bearer '+token}:{}});
    if(mySeq!==mgrLoadSeq) return; // otra carga más reciente desplazó a esta
    if(!res.ok){ if(grid) grid.innerHTML='<div style="color:#ef4444">Error cargando imágenes</div>'; return; }
    const p=await res.json(); const imgs=(p && p.Images)||[];
    if(mySeq!==mgrLoadSeq) return;
    if(!imgs.length){ if(emptyLbl) emptyLbl.style.display=''; return; }
    imgs.forEach(im=>{
      if(!im || !im.URL || !im.ID) return;
      const card=document.createElement('div'); card.className='img-card'; card.dataset.id=im.ID; card.style.position='relative'; card.style.width='92px'; card.style.height='92px'; card.style.cursor='pointer';
      const img=document.createElement('img'); img.src=im.URL; img.alt=im.Alt||''; img.loading='lazy'; img.style.width='100%'; img.style.height='100%'; img.style.objectFit='cover'; img.style.borderRadius='10px'; img.style.border='1px solid #223140'; img.onerror=()=>{ card.remove(); };
      const del=document.createElement('button'); del.textContent='✖'; del.className='icon-btn danger'; del.title='Eliminar'; del.style.position='absolute'; del.style.top='-6px'; del.style.right='-6px'; del.style.background='#ef4444'; del.style.color='#fff'; del.style.borderRadius='50%'; del.style.width='22px'; del.style.height='22px'; del.style.display='grid'; del.style.placeItems='center'; del.style.fontSize='12px'; del.style.cursor='pointer';
      const sel=document.createElement('div'); sel.textContent='✓'; sel.setAttribute('aria-hidden','true'); sel.style.position='absolute'; sel.style.left='-6px'; sel.style.top='-6px'; sel.style.width='20px'; sel.style.height='20px'; sel.style.borderRadius='50%'; sel.style.display='grid'; sel.style.placeItems='center'; sel.style.fontSize='12px'; sel.style.background='#10b981'; sel.style.color='#0b1520'; sel.style.border='1px solid #0f2b3d'; sel.style.boxShadow='0 0 0 2px #0b1520'; sel.style.opacity='0'; sel.style.transition='opacity .15s ease';
      del.addEventListener('click', async (ev)=>{ ev.stopPropagation(); if(!token){ alert('Sesión no válida'); return; } if(!confirm('Eliminar esta imagen?')) return; const res=await fetch('/api/product_images/'+encodeURIComponent(im.ID),{method:'DELETE', headers:{Authorization:'Bearer '+token}}); if(res.ok){ mgrSelected.delete(im.ID); mgrUpdateSel(); card.remove(); } else { alert('Error eliminando'); } });
      card.addEventListener('click', ()=>{ const id=im.ID; if(mgrSelected.has(id)){ mgrSelected.delete(id); card.style.outline=''; sel.style.opacity='0'; } else { mgrSelected.add(id); card.style.outline='2px solid #10b981'; sel.style.opacity='1'; } mgrUpdateSel(); });
      card.appendChild(img); card.appendChild(del); card.appendChild(sel); if(grid) grid.appendChild(card);
    });
  }
  function openMgr(){ const slug=(fSlug&&fSlug.value.trim())||''; if(!slug){ alert('Seleccioná un producto primero'); return; } mgrOpen(); }
  if(btnManage){ btnManage.addEventListener('click', e=>{ e.preventDefault(); openMgr(); }); }
  if(btnManageAlt){ btnManageAlt.addEventListener('click', e=>{ e.preventDefault(); openMgr(); }); }
  if(mgrClose) mgrClose.addEventListener('click', mgrCloseFn);
  if(mgrClose2) mgrClose2.addEventListener('click', mgrCloseFn);
  if(mgrDelBtn){ mgrDelBtn.addEventListener('click', async (e)=>{ e.preventDefault(); if(mgrSelected.size===0) return; if(!token){ alert('Sesión no válida'); return; } if(!confirm('Eliminar imágenes seleccionadas?')) return; const ids=[...mgrSelected]; await Promise.all(ids.map(id=>fetch('/api/product_images/'+encodeURIComponent(id),{method:'DELETE', headers:{Authorization:'Bearer '+token}}))); ids.forEach(id=>{ const el=grid && grid.querySelector('.img-card[data-id="'+CSS.escape(id)+'"]'); if(el) el.remove(); }); mgrSelected.clear(); mgrUpdateSel(); await reloadProductAndSync(); }); }

  // Bulk inline price editing
  const bulkSaveBtn=document.getElementById('bulkSaveBtn');
  const pendingPrices=new Map();

  function bulkUpdateBtn(){
    if(!bulkSaveBtn) return;
    const n=pendingPrices.size;
    if(n===0){ bulkSaveBtn.style.display='none'; return; }
    bulkSaveBtn.textContent='Guardar '+n+' cambio'+(n>1?'s':'');
    bulkSaveBtn.style.display='block';
  }

  if(tbl){
    tbl.addEventListener('input', function(e){
      const inp=e.target.closest('.inline-price');
      if(!inp) return;
      const slug=inp.getAttribute('data-slug');
      const field=inp.getAttribute('data-field');
      const original=inp.getAttribute('data-original');
      const current=parseFloat(inp.value);
      const orig=parseFloat(original);
      const changed=!isNaN(current) && (isNaN(orig) ? current!==0 : Math.abs(current-orig)>0.001);
      if(changed){
        inp.classList.add('changed');
        if(!pendingPrices.has(slug)) pendingPrices.set(slug,{});
        pendingPrices.get(slug)[field]=current;
      } else {
        inp.classList.remove('changed');
        if(pendingPrices.has(slug)){
          delete pendingPrices.get(slug)[field];
          if(Object.keys(pendingPrices.get(slug)).length===0) pendingPrices.delete(slug);
        }
      }
      bulkUpdateBtn();
    });
  }

  if(bulkSaveBtn){
    bulkSaveBtn.addEventListener('click', async function(){
      if(pendingPrices.size===0) return;
      if(!token){ showToast('Sesión no válida','error'); return; }
      const items=[];
      pendingPrices.forEach(function(fields,slug){
        const item={slug:slug};
        if('base_price' in fields) item.base_price=fields.base_price;
        if('gross_price' in fields) item.gross_price=fields.gross_price;
        if('profit' in fields) item.profit=fields.profit;
        items.push(item);
      });
      bulkSaveBtn.disabled=true;
      bulkSaveBtn.textContent='Guardando...';
      try{
        const res=await fetch('/api/products/bulk-prices',{
          method:'PATCH',
          headers:Object.assign({'Content-Type':'application/json'},token?{Authorization:'Bearer '+token}:{}),
          body:JSON.stringify(items)
        });
        if(!res.ok){
          const txt=await res.text();
          showToast('Error: '+txt,'error');
          return;
        }
        document.querySelectorAll('.inline-price.changed').forEach(function(inp){
          inp.setAttribute('data-original',parseFloat(inp.value).toFixed(2));
          inp.classList.remove('changed');
        });
        pendingPrices.clear();
        bulkUpdateBtn();
        showToast('Precios actualizados','success');
      } catch(err){
        showToast('Error de red','error');
      } finally{
        bulkSaveBtn.disabled=false;
        bulkUpdateBtn();
      }
    });
  }
})();

// Admin: calculadora de costos
(function(){
  const form=document.getElementById('costCalcForm'); if(!form) return;
  const statusEl=document.getElementById('ccStatus');
  const resultEl=document.getElementById('costCalcResult');
  // Tooltips ayuda (click toggle)
  form.addEventListener('click',e=>{
    const ico=e.target.closest('.help-icon');
    if(!ico) return;
    e.preventDefault();
    const msg=ico.getAttribute('data-help')||'';
    if(!msg) return;
    let tip=ico.nextElementSibling && ico.nextElementSibling.classList && ico.nextElementSibling.classList.contains('help-tip') ? ico.nextElementSibling : null;
    if(tip){ tip.remove(); return; }
    tip=document.createElement('div');
    tip.className='help-tip';
    tip.textContent=msg;
    tip.style.position='absolute';
    tip.style.zIndex='20';
    tip.style.background='#0F1B2D';
    tip.style.color='#E5F0FF';
    tip.style.border='1px solid #223140';
    tip.style.borderRadius='8px';
    tip.style.padding='8px 10px';
    tip.style.fontSize='12px';
    tip.style.maxWidth='320px';
    tip.style.boxShadow='0 6px 18px rgba(0,0,0,.3)';
    tip.style.marginLeft='8px';
    tip.style.display='inline-block';
    ico.after(tip);
    // Cerrar al clic afuera
    const close=(ev)=>{ if(!tip.contains(ev.target) && ev.target!==ico){ tip.remove(); document.removeEventListener('click',close,true); } };
    setTimeout(()=>document.addEventListener('click',close,true),0);
  });
  function getNum(id){ const el=document.getElementById(id); const v=parseFloat((el&&el.value)||'0'); return isNaN(v)?0:v; }
  function getInt(id){ const el=document.getElementById(id); const v=parseInt((el&&el.value)||'0',10); return isNaN(v)?0:v; }
  function getChk(id){ const el=document.getElementById(id); return !!(el&&el.checked); }
  form.addEventListener('submit', async e=>{
    e.preventDefault();
    const payload={
      price_per_kg: getNum('ccPricePerKg'),
      price_per_kwh: getNum('ccPricePerKwh'),
      power_watts: getNum('ccPowerWatts'),
      machine_wear_hours: 0,
      spare_parts_price: 0,
      error_percent: getNum('ccErrorPct'),
      time_hours: getInt('ccTimeH'),
      time_minutes: getInt('ccTimeM'),
      filament_grams: getNum('ccGrams'),
      supplies_ars: getNum('ccSupplies'),
      margin_multiplier: getNum('ccMarginMult'),
      ml_gross_up: getChk('ccMLGross'),
      ml_fee_percent: getNum('ccMLFeePct'),
      ml_fixed_fee: getNum('ccMLFixed')
    };
    if(statusEl){ statusEl.textContent='Calculando...'; }
    try{
      const res=await fetch('/admin/costs/calculate',{method:'POST', headers:{'Content-Type':'application/json'}, credentials:'same-origin', body:JSON.stringify(payload)});
      if(!res.ok){ const msg=await res.text(); throw new Error(msg||('HTTP '+res.status)); }
      const out=await res.json();
      if(resultEl){
        const rows=[
          ['Material', `$${formatPrice(out.precio_material)}`],
          ['Luz', `$${formatPrice(out.precio_luz)}`],
          ['Error', `$${formatPrice(out.margen_de_error)}`],
          ['Subtotal s/ins.', `$${formatPrice(out.subtotal_sin_insumos)}`],
          ['Total s/ins.', `$${formatPrice(out.total_sin_insumos)}`],
          ['Insumos', `$${formatPrice(out.insumos)}`],
          ['Total a cobrar', `$${formatPrice(out.total_a_cobrar)}`],
          ['Precio ML', `$${formatPrice(out.precio_mercadolibre)}`],
          ['Horas', `${out.horas.toFixed(2)} h`],
          ['Filamento', `${out.filamento_kg.toFixed(2)} kg`],
        ];
        resultEl.innerHTML = '<div class="table clean"></div><div style="margin-top:10px;font-size:13px;color:var(--muted)">Multiplicador usado: '+out.margin_multiplier_used+'</div>';
        const tbl=resultEl.querySelector('.table');
        if(tbl){ rows.forEach(([k,v])=>{ const row=document.createElement('div'); row.className='row between'; row.style.padding='6px 0'; row.innerHTML=`<div>${k}</div><div><strong>${v}</strong></div>`; tbl.appendChild(row); }); }
      }
      if(statusEl){ statusEl.textContent='Listo'; setTimeout(()=>{statusEl.textContent='';},1500); }
    }catch(err){ if(statusEl){ statusEl.textContent='Error: '+(err&&err.message||''); } }
  });
})();

// Búsqueda con autocompletado
(function(){
  const searchInput=document.getElementById('searchInput');
  const mobileSearchInput=document.getElementById('mobileSearchInput');
  const searchResults=document.getElementById('searchResults');
  const mobileSearchResults=document.getElementById('mobileSearchResults');
  let debounceTimer=null;
  let currentQuery='';
  function showResults(input,resultsEl,suggestions){
    if(!resultsEl) return;
    if(!suggestions || suggestions.length===0){
      resultsEl.style.display='none';
      return;
    }
    resultsEl.innerHTML='';
    suggestions.forEach(item=>{
      const div=document.createElement('div');
      div.className='search-result-item';
      div.style.cssText='display:flex;align-items:center;gap:10px;padding:10px;cursor:pointer;background:#12202c;border-bottom:1px solid #223140;transition:background .15s';
      div.onmouseenter=()=>{div.style.background='#1b2a38'};
      div.onmouseleave=()=>{div.style.background='#12202c'};
      div.onclick=()=>{window.location.href='/product/'+item.slug};
      if(item.image){
        const img=document.createElement('img');
        img.src=item.image;
        img.alt='';
        img.style.cssText='width:48px;height:48px;object-fit:contain;border-radius:8px;border:1px solid #223140';
        div.appendChild(img);
      }
      const info=document.createElement('div');
      info.style.cssText='flex:1';
      const name=document.createElement('div');
      name.textContent=item.name;
      name.style.cssText='font-weight:600;margin-bottom:2px';
      const meta=document.createElement('div');
      meta.textContent=item.category+' • $'+formatPrice(item.price);
      meta.style.cssText='font-size:12px;color:#94a3b8';
      info.appendChild(name);
      info.appendChild(meta);
      div.appendChild(info);
      resultsEl.appendChild(div);
    });
    resultsEl.style.display='block';
  }
  function fetchSuggestions(query,targetInput,resultsEl){
    if(!query || query.length<3){
      if(resultsEl) resultsEl.style.display='none';
      return;
    }
    if(query===currentQuery) return;
    currentQuery=query;
    fetch('/api/search/suggestions?q='+encodeURIComponent(query),{credentials:'same-origin'})
      .then(res=>res.json())
      .then(data=>showResults(targetInput,resultsEl,data))
      .catch(()=>{});
  }
  function handleInput(input,resultsEl){
    if(!input || !resultsEl) return;
    input.addEventListener('input',()=>{
      const q=input.value.trim();
      clearTimeout(debounceTimer);
      debounceTimer=setTimeout(()=>fetchSuggestions(q,input,resultsEl),300);
    });
    // Cerrar resultados al hacer clic fuera
    document.addEventListener('click',(e)=>{
      if(!input.contains(e.target) && !resultsEl.contains(e.target)){
        resultsEl.style.display='none';
      }
    });
  }
  handleInput(searchInput,searchResults);
  handleInput(mobileSearchInput,mobileSearchResults);
})();

// Admin: Productos Destacados
(function(){
  const featuredList = document.getElementById('featuredList');
  const allProductsList = document.getElementById('allProductsList');
  const productSearch = document.getElementById('productSearch');
  if(!allProductsList) return; // No estamos en la página de destacada
  
  console.log('=== Inicializando Destacada Admin ===');
  const adminTokenEl = document.querySelector('[data-admin-token]');
  const adminToken = adminTokenEl ? adminTokenEl.getAttribute('data-admin-token') : '';
  console.log('AdminToken:', adminToken ? 'Presente' : 'FALTANTE');
  console.log('Elements found:', {featuredList: !!featuredList, allProductsList: !!allProductsList, productSearch: !!productSearch});
  
  // Log productos disponibles
  const items = allProductsList.querySelectorAll('.admin-list-item');
  console.log('Total productos encontrados:', items.length);
  items.forEach((item, idx) => {
    const id = item.getAttribute('data-product-id');
    const name = item.getAttribute('data-product-name');
    if(idx < 5) console.log(`Producto ${idx+1}: ID="${id}", Name="${name}"`);
  });

  function updateFeaturedStatus(){
    if(!allProductsList) return;
    const featured = document.querySelectorAll('#featuredList .admin-list-item');
    const featuredIds = new Set();
    featured.forEach(item => featuredIds.add(item.getAttribute('data-product-id')));
    
    allProductsList.querySelectorAll('.admin-list-item').forEach(item => {
      const productId = item.getAttribute('data-product-id');
      const btn = item.querySelector('button');
      if(featuredIds.has(productId)){
        item.classList.add('featured');
        if(btn) {
          btn.textContent = 'Destacado';
          btn.disabled = true;
          btn.classList.remove('btn-primary');
          btn.classList.add('btn-secondary');
        }
      }
    });
  }

  function addFeatured(productId){
    console.log('Sending request to add featured product:', productId);
    fetch('/api/featured/add', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + adminToken
      },
      body: JSON.stringify({product_id: productId, order: 0})
    }).then(res => {
      console.log('Response status:', res.status);
      if(res.ok){
        location.reload();
      } else {
        res.text().then(text => {
          console.error('Error response:', text);
          alert('Error: ' + text);
        });
      }
    }).catch(err => {
      console.error('Fetch error:', err);
      alert('Error: ' + err);
    });
  }

  function removeFeatured(productId){
    if(!confirm('¿Quitar este producto de destacados?')) return;
    fetch('/api/featured/remove', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + adminToken
      },
      body: JSON.stringify({product_id: productId})
    }).then(res => {
      if(res.ok){
        location.reload();
      } else {
        res.text().then(text => alert('Error: ' + text));
      }
    }).catch(err => alert('Error: ' + err));
  }

  // Event delegation
  if(featuredList){
    featuredList.addEventListener('click', e => {
      const btn = e.target.closest('[data-action="remove"]');
      if(btn){
        const item = btn.closest('.admin-list-item');
        const productId = item.getAttribute('data-product-id');
        removeFeatured(productId);
      }
    });
  }

  if(allProductsList){
    console.log('Registrando event listener en allProductsList');
    allProductsList.addEventListener('click', e => {
      console.log('Click detectado en allProductsList, target:', e.target);
      const btn = e.target.closest('[data-action="add"]');
      console.log('Button encontrado:', btn);
      if(btn) {
        console.log('Button disabled?', btn.disabled);
      }
      if(btn && !btn.disabled){
        const item = btn.closest('.admin-list-item');
        const productId = item.getAttribute('data-product-id');
        console.log('Adding featured product:', productId);
        addFeatured(productId);
      } else {
        console.log('Botón no encontrado o está deshabilitado');
      }
    });

    // Search
    if(productSearch){
      productSearch.addEventListener('input', e => {
        const query = e.target.value.toLowerCase();
        allProductsList.querySelectorAll('.admin-list-item').forEach(item => {
          const name = item.getAttribute('data-product-name').toLowerCase();
          item.style.display = name.includes(query) ? '' : 'none';
        });
      });
    }
  }

  // Update status on load
  console.log('Ejecutando updateFeaturedStatus()');
  updateFeaturedStatus();
  console.log('=== Destacada Admin inicializado correctamente ===');
})();

// Admin: Carrusel
(function(){
  const carouselList = document.getElementById('carouselList');
  const refreshBtn = document.getElementById('refreshCarouselBtn');
  const allProductsList = document.getElementById('allProductsList');
  if(!carouselList || !allProductsList) return;
  
  const adminTokenEl = document.querySelector('[data-admin-token]');
  const adminToken = adminTokenEl ? adminTokenEl.getAttribute('data-admin-token') : '';
  
  // Función para actualizar el estado de productos en el carrusel
  function updateCarouselStatus(){
    const carouselItems = document.querySelectorAll('#carouselList .carousel-item');
    const carouselSlugs = new Set();
    carouselItems.forEach(item => {
      carouselSlugs.add(item.getAttribute('data-slug'));
    });
    
    allProductsList.querySelectorAll('.admin-list-item').forEach(item => {
      const slug = item.getAttribute('data-product-slug');
      const carouselBtn = item.querySelector('[data-action="add-carousel"]');
      if(carouselSlugs.has(slug)){
        item.classList.add('in-carousel');
        if(carouselBtn) {
          carouselBtn.disabled = true;
          carouselBtn.textContent = 'En Carrusel';
        }
      } else {
        item.classList.remove('in-carousel');
        if(carouselBtn) {
          carouselBtn.disabled = false;
          carouselBtn.textContent = 'Añadir al Carrusel';
        }
      }
    });
  }
  
  // Función para refrescar el carrusel
  function refreshCarousel(){
    location.reload();
  }
  
  // Función para agregar producto al carrusel
  function addToCarousel(slug){
    const carouselItems = document.querySelectorAll('#carouselList .carousel-item');
    const carouselSlugs = [];
    carouselItems.forEach(item => {
      const itemSlug = item.getAttribute('data-slug');
      if(itemSlug) carouselSlugs.push(itemSlug);
    });
    
    // Agregar el nuevo slug
    if(carouselSlugs.length >= 5){
      alert('El carrusel ya tiene el máximo de 5 productos');
      return;
    }
    carouselSlugs.push(slug);
    
    fetch('/api/carousel/update', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + adminToken
      },
      body: JSON.stringify({items: carouselSlugs})
    }).then(res => {
      if(res.ok){
        refreshCarousel();
      } else {
        res.text().then(text => {
          console.error('Error response:', text);
          alert('Error: ' + text);
        });
      }
    }).catch(err => {
      console.error('Fetch error:', err);
      alert('Error: ' + err);
    });
  }
  
  // Función para quitar producto del carrusel
  function removeFromCarousel(slug){
    if(!confirm('¿Quitar este producto del carrusel?')) return;
    
    const carouselItems = document.querySelectorAll('#carouselList .carousel-item');
    const carouselSlugs = [];
    carouselItems.forEach(item => {
      const itemSlug = item.getAttribute('data-slug');
      if(itemSlug && itemSlug !== slug) carouselSlugs.push(itemSlug);
    });
    
    fetch('/api/carousel/update', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + adminToken
      },
      body: JSON.stringify({items: carouselSlugs})
    }).then(res => {
      if(res.ok){
        refreshCarousel();
      } else {
        res.text().then(text => alert('Error: ' + text));
      }
    }).catch(err => alert('Error: ' + err));
  }
  
  // Event delegation para carrusel
  if(carouselList){
    carouselList.addEventListener('click', e => {
      const btn = e.target.closest('[data-action="remove-carousel"]');
      if(btn){
        const item = btn.closest('.carousel-item');
        const slug = item.getAttribute('data-slug');
        removeFromCarousel(slug);
      }
    });
  }
  
  // Event delegation para agregar al carrusel
  if(allProductsList){
    allProductsList.addEventListener('click', e => {
      const btn = e.target.closest('[data-action="add-carousel"]');
      if(btn && !btn.disabled){
        const item = btn.closest('.admin-list-item');
        const slug = item.getAttribute('data-product-slug');
        addToCarousel(slug);
      }
    });
  }
  
  // Botón de refrescar carrusel
  if(refreshBtn){
    refreshBtn.addEventListener('click', refreshCarousel);
  }
  
  // Actualizar estado al cargar
  updateCarouselStatus();
})();

// ============================================
// CART MOBILE OPTIMIZADO - INTERACCIONES
// ============================================

(function(){
  // Elementos principales
  const checkoutForm = document.getElementById('checkoutForm');
  const stickyBottom = document.getElementById('cartStickyBottom');
  const stickyToggle = document.getElementById('cartStickyToggle');
  const stickyExpanded = document.getElementById('cartStickyExpanded');
  
  if(!checkoutForm) return; // No estamos en la página del carrito
  
  // ============ VALIDACIÓN Y HELPERS ============
  
  // Validar email
  function isValidEmail(email) {
    const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return re.test(email);
  }
  
  // Formatear teléfono mientras escribe
  function formatPhoneNumber(value) {
    const numbers = value.replace(/\D/g, '');
    return numbers;
  }
  
  // Mostrar error en campo
  function showFieldError(field, message) {
    if(!field) return;
    
    // Remover error previo si existe
    clearFieldError(field);
    
    field.classList.add('field-invalid');
    
    const errorDiv = document.createElement('div');
    errorDiv.className = 'field-error';
    errorDiv.textContent = message;
    errorDiv.setAttribute('data-error-for', field.name);
    
    field.parentNode.insertBefore(errorDiv, field.nextSibling);
    
    // Limpiar error al escribir
    field.addEventListener('input', () => clearFieldError(field), { once: true });
  }
  
  // Limpiar error de campo
  function clearFieldError(field) {
    if(!field) return;
    field.classList.remove('field-invalid');
    field.classList.remove('field-valid');
    
    const errorDiv = field.parentNode.querySelector(`[data-error-for="${field.name}"]`);
    if(errorDiv) errorDiv.remove();
  }
  
  // Marcar campo como válido
  function markFieldValid(field) {
    if(!field) return;
    clearFieldError(field);
    field.classList.add('field-valid');
  }
  
  // ============ AUTO-FILL CON LOCALSTORAGE ============
  
  function loadFormData() {
    try {
      const saved = localStorage.getItem('checkoutData');
      if(saved) {
        const data = JSON.parse(saved);
        const nameInput = document.querySelector('input[name="name"]');
        const emailInput = document.querySelector('input[name="email"]');
        const phoneInput = document.querySelector('input[name="phone"]');
        
        if(nameInput && data.name) nameInput.value = data.name;
        if(emailInput && data.email) emailInput.value = data.email;
        if(phoneInput && data.phone) phoneInput.value = data.phone;
      }
    } catch(e) {
      console.error('Error loading form data:', e);
    }
  }
  
  function saveFormData() {
    try {
      const nameInput = document.querySelector('input[name="name"]');
      const emailInput = document.querySelector('input[name="email"]');
      const phoneInput = document.querySelector('input[name="phone"]');
      
      const formData = {
        name: nameInput ? nameInput.value : '',
        email: emailInput ? emailInput.value : '',
        phone: phoneInput ? phoneInput.value : ''
      };
      
      localStorage.setItem('checkoutData', JSON.stringify(formData));
    } catch(e) {
      console.error('Error saving form data:', e);
    }
  }
  
  // Cargar datos guardados al inicio
  loadFormData();
  
  // ============ BARRA DE PROGRESO DINÁMICA ============
  
  function updateProgressBar(step) {
    const progressFill = document.getElementById('progressBarFill');
    const currentStepEl = document.getElementById('currentStep');
    const progressPercentEl = document.getElementById('progressPercent');
    
    const percentages = {
      1: 0,
      2: 33,
      3: 66,
      4: 100
    };
    
    const percent = percentages[step] || 0;
    
    if(progressFill) progressFill.style.width = percent + '%';
    if(currentStepEl) currentStepEl.textContent = step;
    if(progressPercentEl) progressPercentEl.textContent = percent;
  }
  
  // ============ NAVEGACIÓN ENTRE SECCIONES ============
  
  const sectionHeaders = document.querySelectorAll('.checkout-section-header');
  const sections = ['contact', 'shipping', 'payment'];
  let completedSections = new Set();
  
  function collapseSection(sectionName) {
    const header = document.querySelector(`[data-section-name="${sectionName}"]`);
    const content = header ? header.nextElementSibling : null;
    
    if(header) header.classList.remove('active');
    if(content) content.classList.remove('active');
  }
  
  function openSection(sectionName) {
    const header = document.querySelector(`[data-section-name="${sectionName}"]`);
    const content = header ? header.nextElementSibling : null;
    
    if(header && !header.classList.contains('disabled')) {
      header.classList.add('active');
      if(content) content.classList.add('active');
    }
  }
  
  function enableSection(sectionName) {
    const header = document.querySelector(`[data-section-name="${sectionName}"]`);
    if(header) header.classList.remove('disabled');
  }
  
  function markSectionComplete(sectionName) {
    const header = document.querySelector(`[data-section-name="${sectionName}"]`);
    if(header) {
      header.classList.add('completed');
      completedSections.add(sectionName);
      
      // Actualizar estado de botones de submit
      setTimeout(() => {
        if(typeof updateSubmitButtons === 'function') {
          updateSubmitButtons();
        }
      }, 100);
    }
  }
  
  function scrollToSection(sectionName) {
    const header = document.querySelector(`[data-section-name="${sectionName}"]`);
    if(header) {
      const offset = 100;
      const y = header.getBoundingClientRect().top + window.pageYOffset - offset;
      window.scrollTo({ top: y, behavior: 'smooth' });
    }
  }
  
  function focusFirstInput(sectionName) {
    const header = document.querySelector(`[data-section-name="${sectionName}"]`);
    const content = header ? header.nextElementSibling : null;
    if(content) {
      const firstInput = content.querySelector('input:not([type="radio"]):not([type="hidden"]), select');
      if(firstInput) {
        setTimeout(() => firstInput.focus(), 400);
      }
    }
  }
  
  function moveToNextSection(currentSection) {
    const currentIndex = sections.indexOf(currentSection);
    const nextSection = sections[currentIndex + 1];
    
    if(nextSection) {
      collapseSection(currentSection);
      markSectionComplete(currentSection);
      
      setTimeout(() => {
        enableSection(nextSection);
        openSection(nextSection);
        scrollToSection(nextSection);
        focusFirstInput(nextSection);
        
        // Actualizar barra de progreso
        updateProgressBar(currentIndex + 2);
      }, 300);
    } else {
      // Última sección completada
      markSectionComplete(currentSection);
      updateProgressBar(4);
    }
  }
  
  // ============ VALIDACIÓN PROGRESIVA ============
  
  function validateContactSection() {
    const nameInput = document.querySelector('input[name="name"]');
    const emailInput = document.querySelector('input[name="email"]');
    const phoneInput = document.querySelector('input[name="phone"]');
    
    let isValid = true;
    
    // Validar nombre
    if(!nameInput || !nameInput.value.trim()) {
      if(nameInput) showFieldError(nameInput, 'El nombre es requerido');
      isValid = false;
    } else {
      if(nameInput) markFieldValid(nameInput);
    }
    
    // Validar email
    if(!emailInput || !emailInput.value.trim()) {
      if(emailInput) showFieldError(emailInput, 'El email es requerido');
      isValid = false;
    } else if(!isValidEmail(emailInput.value.trim())) {
      showFieldError(emailInput, 'Ingresá un email válido');
      isValid = false;
    } else {
      markFieldValid(emailInput);
    }
    
    // Validar teléfono
    if(!phoneInput || !phoneInput.value.trim()) {
      if(phoneInput) showFieldError(phoneInput, 'El teléfono es requerido');
      isValid = false;
    } else if(phoneInput.value.trim().length < 8) {
      showFieldError(phoneInput, 'Ingresá un teléfono válido (mínimo 8 dígitos)');
      isValid = false;
    } else {
      markFieldValid(phoneInput);
    }
    
    if(isValid) {
      saveFormData();
    }
    
    return isValid;
  }
  
  function validateShippingSection() {
    const shippingMethod = document.querySelector('input[name="shipping"]:checked');
    
    if(!shippingMethod) {
      showToast('Seleccioná un método de envío', 'error');
      return false;
    }
    
    // Validar campos condicionales de cadete
    if(shippingMethod.value === 'cadete') {
      const addressCadete = document.querySelector('input[name="address_cadete"]');
      if(!addressCadete || !addressCadete.value.trim()) {
        if(addressCadete) showFieldError(addressCadete, 'La dirección es requerida');
        return false;
      } else {
        markFieldValid(addressCadete);
      }
    }
    
    // Validar campos condicionales de envío
    if(shippingMethod.value === 'envio') {
      const provinceSelect = document.querySelector('select[name="province"]');
      const addressEnvio = document.querySelector('input[name="address_envio"]');
      const postalCode = document.querySelector('input[name="postal_code"]');
      const dni = document.querySelector('input[name="dni"]');
      
      let isValid = true;
      
      if(!provinceSelect || !provinceSelect.value) {
        if(provinceSelect) showFieldError(provinceSelect, 'Seleccioná una provincia');
        isValid = false;
      } else {
        if(provinceSelect) markFieldValid(provinceSelect);
      }
      
      if(!addressEnvio || !addressEnvio.value.trim()) {
        if(addressEnvio) showFieldError(addressEnvio, 'La dirección es requerida');
        isValid = false;
      } else {
        if(addressEnvio) markFieldValid(addressEnvio);
      }
      
      if(!postalCode || !postalCode.value.trim()) {
        if(postalCode) showFieldError(postalCode, 'El código postal es requerido');
        isValid = false;
      } else {
        if(postalCode) markFieldValid(postalCode);
      }
      
      if(!dni || !dni.value.trim()) {
        if(dni) showFieldError(dni, 'El DNI es requerido');
        isValid = false;
      } else {
        if(dni) markFieldValid(dni);
      }
      
      return isValid;
    }
    
    return true;
  }
  
  function validatePaymentSection() {
    const paymentMethod = document.querySelector('input[name="payment_method"]:checked');
    
    if(!paymentMethod) {
      showToast('Seleccioná un método de pago', 'error');
      return false;
    }
    
    return true;
  }
  
  // ============ VALIDACIÓN EN TIEMPO REAL ============
  
  // Validación de contacto en tiempo real
  const nameInput = document.querySelector('input[name="name"]');
  const emailInput = document.querySelector('input[name="email"]');
  const phoneInput = document.querySelector('input[name="phone"]');
  
  if(nameInput) {
    nameInput.addEventListener('blur', () => {
      if(nameInput.value.trim()) {
        markFieldValid(nameInput);
      }
      
      // Auto-avanzar si todos los campos están completos
      if(validateContactSection()) {
        setTimeout(() => {
          if(document.querySelector('[data-section-name="contact"]').classList.contains('active')) {
            moveToNextSection('contact');
          }
        }, 500);
      }
    });
  }
  
  if(emailInput) {
    emailInput.addEventListener('blur', () => {
      const email = emailInput.value.trim();
      if(email && isValidEmail(email)) {
        markFieldValid(emailInput);
      }
      
      // Auto-avanzar si todos los campos están completos
      if(validateContactSection()) {
        setTimeout(() => {
          if(document.querySelector('[data-section-name="contact"]').classList.contains('active')) {
            moveToNextSection('contact');
          }
        }, 500);
      }
    });
    
    // Validación en tiempo real
    emailInput.addEventListener('input', () => {
      const email = emailInput.value.trim();
      if(email.length > 3 && !isValidEmail(email)) {
        emailInput.classList.add('field-invalid');
      } else if(email && isValidEmail(email)) {
        emailInput.classList.remove('field-invalid');
        emailInput.classList.add('field-valid');
      } else {
        emailInput.classList.remove('field-invalid');
        emailInput.classList.remove('field-valid');
      }
    });
  }
  
  if(phoneInput) {
    // Formatear teléfono
    phoneInput.addEventListener('input', (e) => {
      const formatted = formatPhoneNumber(e.target.value);
      e.target.value = formatted;
      
      if(formatted.length >= 8) {
        markFieldValid(phoneInput);
      }
    });
    
    phoneInput.addEventListener('blur', () => {
      // Auto-avanzar si todos los campos están completos
      if(validateContactSection()) {
        setTimeout(() => {
          if(document.querySelector('[data-section-name="contact"]').classList.contains('active')) {
            moveToNextSection('contact');
          }
        }, 500);
      }
    });
  }
  
  // ============ ACORDEÓN CHECKOUT CON VALIDACIÓN ============
  
  sectionHeaders.forEach(header => {
    header.addEventListener('click', function(){
      // No permitir clic en secciones deshabilitadas
      if(this.classList.contains('disabled')) {
        showToast('Completá la sección anterior primero', 'warning');
        return;
      }
      
      const section = this.getAttribute('data-toggle');
      const content = this.nextElementSibling;
      const isActive = this.classList.contains('active');
      
      if(!isActive) {
        // Cerrar todas las secciones
        sectionHeaders.forEach(h => {
          h.classList.remove('active');
          if(h.nextElementSibling) h.nextElementSibling.classList.remove('active');
        });
        
        // Abrir la sección clickeada
        this.classList.add('active');
        content.classList.add('active');
        
        // Auto-focus en el primer input
        setTimeout(() => {
          const firstInput = content.querySelector('input:not([type="radio"]):not([type="hidden"]), select');
          if(firstInput) firstInput.focus();
        }, 400);
      }
    });
  });
  
  // ============ AUTO-AVANCE EN MÉTODOS DE ENVÍO Y PAGO ============
  
  const shippingRadios = document.querySelectorAll('input[name="shipping"]');
  const cadeteGroup = document.getElementById('cadeteGroup');
  const envioGroup = document.getElementById('envioGroup');
  
  shippingRadios.forEach(radio => {
    radio.addEventListener('change', function(){
      if(cadeteGroup) cadeteGroup.style.display = 'none';
      if(envioGroup) envioGroup.style.display = 'none';
      
      if(this.value === 'cadete' && cadeteGroup) {
        cadeteGroup.style.display = 'flex';
        // Focus en el primer campo
        setTimeout(() => {
          const firstInput = cadeteGroup.querySelector('input');
          if(firstInput) firstInput.focus();
        }, 100);
      } else if(this.value === 'envio' && envioGroup) {
        envioGroup.style.display = 'flex';
        // Focus en el primer campo
        setTimeout(() => {
          const firstSelect = envioGroup.querySelector('select');
          if(firstSelect) firstSelect.focus();
        }, 100);
      } else if(this.value === 'retiro') {
        // Si es retiro, auto-validar y avanzar después de un breve delay
        setTimeout(() => {
          if(validateShippingSection()) {
            setTimeout(() => {
              if(document.querySelector('[data-section-name="shipping"]').classList.contains('active')) {
                moveToNextSection('shipping');
              }
            }, 500);
          }
        }, 300);
      }
      
      updateTotals();
    });
  });
  
  // Validar automáticamente cuando se completan campos de cadete
  if(cadeteGroup) {
    const addressCadete = cadeteGroup.querySelector('input[name="address_cadete"]');
    if(addressCadete) {
      addressCadete.addEventListener('blur', () => {
        if(addressCadete.value.trim()) {
          markFieldValid(addressCadete);
          // Auto-avanzar
          setTimeout(() => {
            if(validateShippingSection()) {
              setTimeout(() => {
                if(document.querySelector('[data-section-name="shipping"]').classList.contains('active')) {
                  moveToNextSection('shipping');
                }
              }, 500);
            }
          }, 300);
        }
      });
    }
  }
  
  // Validar automáticamente cuando se completan campos de envío
  if(envioGroup) {
    const dniInput = envioGroup.querySelector('input[name="dni"]');
    if(dniInput) {
      dniInput.addEventListener('blur', () => {
        if(dniInput.value.trim()) {
          markFieldValid(dniInput);
          // Auto-avanzar si todo está completo
          setTimeout(() => {
            if(validateShippingSection()) {
              setTimeout(() => {
                if(document.querySelector('[data-section-name="shipping"]').classList.contains('active')) {
                  moveToNextSection('shipping');
                }
              }, 500);
            }
          }, 300);
        }
      });
    }
  }
  
  // Ocultar mensajes de ayuda al seleccionar
  shippingRadios.forEach(radio => {
    radio.addEventListener('change', function(){
      const shippingSection = document.querySelector('[data-section="shipping"] .checkout-section-content');
      const helpText = shippingSection ? shippingSection.querySelector('.section-help-text') : null;
      if(helpText) {
        helpText.style.opacity = '0';
        helpText.style.transform = 'translateY(-5px)';
        setTimeout(() => {
          if(helpText.parentNode) helpText.parentNode.removeChild(helpText);
        }, 300);
      }
    }, { once: true });
  });
  
  // Auto-avance en método de pago
  const paymentRadios = document.querySelectorAll('input[name="payment_method"]');
  paymentRadios.forEach(radio => {
    radio.addEventListener('change', function(){
      // Ocultar mensaje de ayuda
      const paymentSection = document.querySelector('[data-section="payment"] .checkout-section-content');
      const helpText = paymentSection ? paymentSection.querySelector('.section-help-text') : null;
      if(helpText) {
        helpText.style.opacity = '0';
        helpText.style.transform = 'translateY(-5px)';
        setTimeout(() => {
          if(helpText.parentNode) helpText.parentNode.removeChild(helpText);
        }, 300);
      }
      
      updateTotals();
      
      // Auto-validar y completar
      setTimeout(() => {
        if(validatePaymentSection()) {
          setTimeout(() => {
            if(document.querySelector('[data-section-name="payment"]').classList.contains('active')) {
              markSectionComplete('payment');
              updateProgressBar(4);
              
              // Cerrar la sección de pago
              collapseSection('payment');
              
              // Scroll al botón de confirmar con animación
              setTimeout(() => {
                const submitBtn = document.querySelector('.cart-summary-cta, .cart-sticky-cta');
                if(submitBtn) {
                  submitBtn.scrollIntoView({ behavior: 'smooth', block: 'center' });
                  // Pulsar el botón para llamar la atención
                  submitBtn.style.animation = 'ctaPulse 2s ease-in-out infinite';
                }
              }, 400);
            }
          }, 500);
        }
      }, 300);
    });
  });
  
  // ============ TOGGLE CUPÓN COLAPSABLE ============
  
  const couponToggleLink = document.getElementById('couponToggleLink');
  const couponFieldWrapper = document.getElementById('couponFieldWrapper');
  
  if(couponToggleLink && couponFieldWrapper) {
    couponToggleLink.addEventListener('click', function() {
      const isHidden = couponFieldWrapper.style.display === 'none';
      
      if(isHidden) {
        couponFieldWrapper.style.display = 'block';
        couponToggleLink.classList.add('active');
        
        // Focus en el input
        setTimeout(() => {
          const couponInput = couponFieldWrapper.querySelector('#coupon_code');
          if(couponInput) couponInput.focus();
        }, 100);
      } else {
        couponFieldWrapper.style.display = 'none';
        couponToggleLink.classList.remove('active');
      }
    });
  }
  
  // ============ STICKY BOTTOM BAR ============
  if(stickyToggle && stickyBottom) {
    stickyToggle.addEventListener('click', function(){
      stickyBottom.classList.toggle('expanded');
    });
  }
  
  // ============ CÁLCULO DE TOTALES ============
  function updateTotals() {
    const pcData = document.getElementById('pcData');
    const provinceSelect = document.getElementById('provinceSelect');
    const shippingMethod = document.querySelector('input[name="shipping"]:checked');
    const paymentMethod = document.querySelector('input[name="payment_method"]:checked');
    
    // Obtener subtotal (suma de productos)
    let subtotal = 0;
    const subtotalEl = document.getElementById('subtotalVal');
    if(subtotalEl) {
      const text = subtotalEl.textContent.replace(/[$.,]/g, '');
      subtotal = parseFloat(text) || 0;
    }
    
    // Calcular costo de envío
    let shipCost = 0;
    if(shippingMethod) {
      if(shippingMethod.value === 'cadete') {
        shipCost = 5000;
      } else if(shippingMethod.value === 'envio' && provinceSelect && pcData) {
        const province = provinceSelect.value;
        const provinceCostEl = pcData.querySelector(`[data-prov="${province}"]`);
        if(provinceCostEl) {
          shipCost = parseFloat(provinceCostEl.getAttribute('data-cost')) || 0;
        }
      }
    }
    
    // Obtener descuento del cupón (si está validado)
    let discount = window.appliedCouponDiscount || 0;
    const discountRow = document.getElementById('discountRow');
    const stickyDiscountRow = document.getElementById('stickyDiscountRow');
    
    if(discount > 0) {
      if(discountRow) discountRow.style.display = 'flex';
      if(stickyDiscountRow) stickyDiscountRow.style.display = 'flex';
    } else {
      if(discountRow) discountRow.style.display = 'none';
      if(stickyDiscountRow) stickyDiscountRow.style.display = 'none';
    }
    
    // Calcular total final
    const total = subtotal + shipCost - discount;
    
    // Actualizar todos los displays
    const updateDisplay = (id, value) => {
      const el = document.getElementById(id);
      if(el) el.textContent = '$' + formatPrice(value);
    };
    
    updateDisplay('shipCost', shipCost);
    updateDisplay('discount', discount);
    updateDisplay('finalTotal', total);
    
    // Sticky bar mobile
    updateDisplay('stickyTotal', total);
    updateDisplay('stickySubtotal', subtotal);
    updateDisplay('stickyShipCost', shipCost);
    updateDisplay('stickyDiscount', discount);
  }
  
  // Escuchar cambios en provincia
  const provinceSelect = document.getElementById('provinceSelect');
  if(provinceSelect) {
    provinceSelect.addEventListener('change', updateTotals);
  }
  
  // Calcular totales al cargar
  updateTotals();
  
  // Actualizar barra de progreso inicial
  updateProgressBar(1);
  
  // ============ VALIDACIÓN DE CUPONES ============
  const validateCouponBtn = document.getElementById('validate-coupon-btn');
  const couponCodeInput = document.getElementById('coupon_code');
  const couponMessage = document.getElementById('coupon-message');
  
  // Variable global para almacenar el descuento del cupón
  window.appliedCouponDiscount = 0;
  
  if(validateCouponBtn && couponCodeInput && couponMessage) {
    validateCouponBtn.addEventListener('click', function() {
      const code = couponCodeInput.value.trim();
      
      if(!code) {
        couponMessage.textContent = 'Ingresá un código de cupón';
        couponMessage.className = 'coupon-message coupon-error';
        couponMessage.style.display = 'block';
        return;
      }
      
      // Obtener email del usuario
      const emailInput = document.querySelector('input[name="email"]');
      const email = emailInput ? emailInput.value.trim() : '';
      
      if(!email) {
        couponMessage.textContent = 'Ingresá tu email primero';
        couponMessage.className = 'coupon-message coupon-error';
        couponMessage.style.display = 'block';
        return;
      }
      
      // Calcular subtotal con envío
      let subtotal = 0;
      const subtotalEl = document.getElementById('subtotalVal');
      if(subtotalEl) {
        const text = subtotalEl.textContent.replace(/[$.,]/g, '');
        subtotal = parseFloat(text) || 0;
      }
      
      // Calcular costo de envío
      const shippingMethod = document.querySelector('input[name="shipping"]:checked');
      let shipCost = 0;
      if(shippingMethod) {
        if(shippingMethod.value === 'cadete') {
          shipCost = 5000;
        } else if(shippingMethod.value === 'envio') {
          const provinceSelect = document.getElementById('provinceSelect');
          const pcData = document.getElementById('pcData');
          if(provinceSelect && pcData) {
            const province = provinceSelect.value;
            const provinceCostEl = pcData.querySelector(`[data-prov="${province}"]`);
            if(provinceCostEl) {
              shipCost = parseFloat(provinceCostEl.getAttribute('data-cost')) || 0;
            }
          }
        }
      }
      
      const totalWithShip = subtotal + shipCost;
      
      // Deshabilitar botón mientras valida
      validateCouponBtn.disabled = true;
      validateCouponBtn.textContent = 'Validando...';
      
      // Llamar a la API
      fetch(`/api/validate-coupon?code=${encodeURIComponent(code)}&email=${encodeURIComponent(email)}&subtotal=${totalWithShip}`)
        .then(res => res.json())
        .then(data => {
          if(data.valid) {
            // Cupón válido
            window.appliedCouponDiscount = data.discount;
            couponMessage.textContent = data.message;
            couponMessage.className = 'coupon-message coupon-success';
            couponMessage.style.display = 'block';
            
            // Actualizar label del descuento
            const discountLabel = document.getElementById('discountLabel');
            if(discountLabel) {
              discountLabel.textContent = `Descuento (${code})`;
            }
            
            // Actualizar totales
            updateTotals();
            
            // Cambiar botón a "Aplicado"
            validateCouponBtn.textContent = '✓ Aplicado';
            validateCouponBtn.disabled = true;
            couponCodeInput.readOnly = true;
            
            showToast('Cupón aplicado correctamente', 'success');
          } else {
            // Cupón inválido
            window.appliedCouponDiscount = 0;
            couponMessage.textContent = data.message;
            couponMessage.className = 'coupon-message coupon-error';
            couponMessage.style.display = 'block';
            updateTotals();
            
            validateCouponBtn.disabled = false;
            validateCouponBtn.textContent = 'Validar';
          }
        })
        .catch(err => {
          console.error('Error validating coupon:', err);
          couponMessage.textContent = 'Error al validar el cupón';
          couponMessage.className = 'coupon-message coupon-error';
          couponMessage.style.display = 'block';
          
          validateCouponBtn.disabled = false;
          validateCouponBtn.textContent = 'Validar';
        });
    });
    
    // Permitir validar con Enter
    couponCodeInput.addEventListener('keypress', function(e) {
      if(e.key === 'Enter') {
        e.preventDefault();
        validateCouponBtn.click();
      }
    });
  }
  
  // ============ MICRO-INTERACCIONES ============
  
  // Animación al cambiar cantidad
  const qtyForms = document.querySelectorAll('.cart-qty-form');
  qtyForms.forEach(form => {
    form.addEventListener('submit', function(e){
      const card = this.closest('.cart-product-card');
      if(card) {
        card.classList.add('added');
        setTimeout(() => card.classList.remove('added'), 400);
      }
      
      const btn = e.submitter;
      if(btn && btn.classList.contains('cart-qty-btn')) {
        btn.classList.add('shake');
        setTimeout(() => btn.classList.remove('shake'), 300);
      }
    });
  });
  
  // Confirmación antes de eliminar (solo mobile)
  if(window.innerWidth < 768) {
    const removeForms = document.querySelectorAll('.cart-remove-form');
    removeForms.forEach(form => {
      form.addEventListener('submit', function(e){
        e.preventDefault();
        const card = this.closest('.cart-product-card');
        const productName = card ? card.querySelector('.cart-product-name').textContent : 'este producto';
        
        if(confirm(`¿Eliminar ${productName} del carrito?`)) {
          this.submit();
        }
      });
    });
  }
  
  // ============ CONTROL DE BOTONES DE SUBMIT ============
  
  const submitBtnDesktop = document.querySelector('.cart-summary-cta');
  const submitBtnMobile = document.querySelector('.cart-sticky-cta');
  let isSubmitting = false; // Flag para prevenir múltiples submits
  
  // Función para verificar si todas las secciones están completadas
  function areAllSectionsComplete() {
    // Verificar que las 3 secciones obligatorias estén completadas
    const requiredSections = ['contact', 'shipping', 'payment'];
    return requiredSections.every(section => completedSections.has(section));
  }
  
  // Función para actualizar estado de botones
  function updateSubmitButtons() {
    const allComplete = areAllSectionsComplete();
    
    [submitBtnDesktop, submitBtnMobile].forEach(btn => {
      if(!btn) return;
      
      if(allComplete && !isSubmitting) {
        btn.disabled = false;
        btn.style.opacity = '1';
        btn.style.cursor = 'pointer';
        btn.title = '';
      } else if(isSubmitting) {
        btn.disabled = true;
        btn.style.opacity = '0.6';
        btn.style.cursor = 'not-allowed';
        btn.title = 'Procesando pedido...';
      } else {
        btn.disabled = true;
        btn.style.opacity = '0.6';
        btn.style.cursor = 'not-allowed';
        btn.title = 'Completá todos los campos requeridos';
      }
    });
  }
  
  // Deshabilitar botones inicialmente
  updateSubmitButtons();
  
  // ============ VALIDACIÓN COMPLETA AL SUBMIT ============
  checkoutForm.addEventListener('submit', function(e){
    e.preventDefault();
    
    // Prevenir múltiples submits
    if(isSubmitting) {
      showToast('Tu pedido ya está siendo procesado', 'warning');
      return false;
    }
    
    // Validar que todas las secciones estén completas
    if(!areAllSectionsComplete()) {
      showToast('Por favor completá todos los campos requeridos', 'error');
      
      // Abrir la primera sección incompleta
      const incompleteSections = ['contact', 'shipping', 'payment'].filter(s => !completedSections.has(s));
      if(incompleteSections.length > 0) {
        const firstIncomplete = incompleteSections[0];
        openSection(firstIncomplete);
        scrollToSection(firstIncomplete);
        focusFirstInput(firstIncomplete);
      }
      
      return false;
    }
    
    // Validar secciones individualmente
    if(!validateContactSection()) {
      openSection('contact');
      scrollToSection('contact');
      showToast('Por favor completá tus datos de contacto correctamente', 'error');
      return false;
    }
    
    if(!validateShippingSection()) {
      openSection('shipping');
      scrollToSection('shipping');
      showToast('Por favor completá los datos de envío correctamente', 'error');
      return false;
    }
    
    if(!validatePaymentSection()) {
      openSection('payment');
      scrollToSection('payment');
      showToast('Por favor seleccioná un método de pago', 'error');
      return false;
    }
    
    const shippingMethod = document.querySelector('input[name="shipping"]:checked');
    
    // Validar campos de cadete
    if(shippingMethod && shippingMethod.value === 'cadete') {
      const addressCadete = document.querySelector('input[name="address_cadete"]');
      if(addressCadete && !addressCadete.value.trim()) {
        showToast('Por favor ingresá la dirección para el cadete', 'error');
        openSection('shipping');
        scrollToSection('shipping');
        addressCadete.focus();
        return false;
      }
    }
    
    // Validar campos de envío
    if(shippingMethod && shippingMethod.value === 'envio') {
      const province = document.querySelector('select[name="province"]');
      const address = document.querySelector('input[name="address_envio"]');
      const postalCode = document.querySelector('input[name="postal_code"]');
      const dni = document.querySelector('input[name="dni"]');
      
      if(!province || !province.value) {
        showToast('Por favor seleccioná una provincia', 'error');
        openSection('shipping');
        scrollToSection('shipping');
        if(province) province.focus();
        return false;
      }
      
      if(!address || !address.value.trim()) {
        showToast('Por favor ingresá la dirección de envío', 'error');
        openSection('shipping');
        scrollToSection('shipping');
        if(address) address.focus();
        return false;
      }
      
      if(!postalCode || !postalCode.value.trim()) {
        showToast('Por favor ingresá el código postal', 'error');
        openSection('shipping');
        scrollToSection('shipping');
        if(postalCode) postalCode.focus();
        return false;
      }
      
      if(!dni || !dni.value.trim()) {
        showToast('Por favor ingresá tu DNI', 'error');
        openSection('shipping');
        scrollToSection('shipping');
        if(dni) dni.focus();
        return false;
      }
    }
    
    // Marcar como enviando
    isSubmitting = true;
    updateSubmitButtons();
    
    // Mostrar loading en ambos botones
    [submitBtnDesktop, submitBtnMobile].forEach(btn => {
      if(btn) {
        btn.disabled = true;
        btn.innerHTML = '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2.5" style="animation:spin 1s linear infinite"><circle cx="12" cy="12" r="10" stroke-opacity="0.25"/><path d="M12 2a10 10 0 0110 10"/></svg><span>Procesando...</span>';
      }
    });
    
    // Guardar datos para futuras compras
    saveFormData();
    
    // Mostrar mensaje de confirmación
    showToast('Procesando tu pedido...', 'info', 2000);
    
    // Enviar el formulario
    this.submit();
  });
  
  // ============ FOMO DINÁMICO ============
  const fomoBadge = document.querySelector('.cart-fomo-badge');
  if(fomoBadge) {
    // Simular cantidad de personas viendo (2-5 personas)
    setInterval(() => {
      const viewers = Math.floor(Math.random() * 4) + 2; // 2-5
      const strong = fomoBadge.querySelector('strong');
      if(strong) strong.textContent = viewers + ' personas';
    }, 15000); // Cada 15 segundos
  }
  
  // ============ SCROLL OPTIMIZATION ============
  let lastScroll = 0;
  let scrollTimeout;
  
  window.addEventListener('scroll', function(){
    clearTimeout(scrollTimeout);
    
    scrollTimeout = setTimeout(() => {
      const currentScroll = window.pageYOffset;
      
      // Auto-colapsar sticky bar si scrolleamos hacia arriba
      if(stickyBottom && stickyBottom.classList.contains('expanded')) {
        if(currentScroll < lastScroll) {
          stickyBottom.classList.remove('expanded');
        }
      }
      
      lastScroll = currentScroll;
    }, 100);
  }, { passive: true });
  
})();

// Animación de loading (spinner)
const style = document.createElement('style');
style.textContent = '@keyframes spin{to{transform:rotate(360deg)}}';
document.head.appendChild(style);

// ============================================
// MODAL MAYORISTA
// ============================================

// Usar delegación de eventos para que funcione incluso si los elementos se cargan después
document.addEventListener('click', function(e){
  const btn = e.target.closest('#btnMayorista');
  if(btn){
    e.preventDefault();
    e.stopPropagation();
    const backdrop = document.getElementById('mayoristaBackdrop');
    if(backdrop){
      backdrop.style.display = 'flex';
      document.body.style.overflow = 'hidden';
    }
  }
  
  const cerrar = e.target.closest('#mayoristaCerrar');
  if(cerrar){
    e.preventDefault();
    e.stopPropagation();
    const backdrop = document.getElementById('mayoristaBackdrop');
    if(backdrop){
      backdrop.style.display = 'none';
      document.body.style.overflow = '';
    }
  }
  
  const backdrop = e.target.closest('#mayoristaBackdrop');
  if(backdrop && e.target === backdrop){
    backdrop.style.display = 'none';
    document.body.style.overflow = '';
  }
});

// Cerrar con tecla Escape
document.addEventListener('keydown', function(e){
  if(e.key === 'Escape'){
    const backdrop = document.getElementById('mayoristaBackdrop');
    if(backdrop && backdrop.style.display === 'flex'){
      backdrop.style.display = 'none';
      document.body.style.overflow = '';
    }
  }
});

// Asegurar que el modal esté oculto al cargar
(function(){
  function hideModal(){
    const backdrop = document.getElementById('mayoristaBackdrop');
    if(backdrop){
      backdrop.style.display = 'none';
    }
  }
  
  if(document.readyState === 'loading'){
    document.addEventListener('DOMContentLoaded', hideModal);
  } else {
    hideModal();
  }
  
  window.addEventListener('load', hideModal);
})();