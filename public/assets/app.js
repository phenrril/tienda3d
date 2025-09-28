// UI scripts unificados (module)
// - Nav drawer
// - Carousel home
// - Modal "Cómo comprar"
// - Products drawer/sheet + load more
// - Carrito: cálculo de envío
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
  document.addEventListener('keydown',e=>{if(e.key==='Escape' && bd && !bd.hidden) close();});
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

// Carrito: cálculo de envío y total (compatible con CSP)
(function(){
  const form=document.getElementById('checkoutForm'); if(!form) return;
  const shipRadios=form.querySelectorAll('input[name="shipping"]');
  const envioGroup=document.getElementById('envioGroup');
  const cadeteGroup=document.getElementById('cadeteGroup');
  const provinceSelect=document.getElementById('provinceSelect');
  const shipCostEl=document.getElementById('shipCost');
  const grandEl=document.getElementById('grandTotal');
  const subtotalEl=document.getElementById('subtotalVal');
  const base=parseFloat(((grandEl && grandEl.textContent) || '').replace(/[^0-9.,]/g,'').replace(',','.'))||0;
  const CADETE_COST=5000;
  const COSTS=(()=>{ const m={}; document.querySelectorAll('#pcData [data-prov]').forEach(n=>{ const k=n.getAttribute('data-prov'); const v=parseFloat(n.getAttribute('data-cost')||'0'); if(k){ m[k]=v; } }); return m; })();
  const phone = form.querySelector('input[name="phone"]');
  const addrCadete = form.querySelector('input[name="address_cadete"]');
  const addrEnvio = form.querySelector('input[name="address_envio"]');
  const postal = form.querySelector('input[name="postal_code"]');
  const dni = form.querySelector('input[name="dni"]');
  function setRequired(el,flag){ if(!el) return; if(flag){el.setAttribute('required','required')} else {el.removeAttribute('required')} }
  function calcCost(){
    let method='retiro'; shipRadios.forEach(r=>{ if(r.checked) method=r.value });
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
    if(shipCostEl) shipCostEl.textContent='$'+cost.toFixed(2);
    const withShip=(base+cost).toFixed(2);
    if(grandEl) grandEl.textContent='$'+withShip;
    if(subtotalEl) subtotalEl.textContent='$'+withShip;
  }
  shipRadios.forEach(r=>r.addEventListener('change',calcCost));
  provinceSelect && provinceSelect.addEventListener('change',calcCost);
  calcCost();
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

// Admin: productos (form + tabla + dropzone) sin inline JS
(function(){
  const form=document.getElementById('prodForm'); if(!form) return;
  const tbl=document.getElementById('prodTable');
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
    if(fSlug) fSlug.value=p.Slug||''; if(fName) fName.value=p.Name||''; if(fCat) fCat.value=p.Category||''; if(fDesc) fDesc.value=p.ShortDesc||''; if(fPrice) fPrice.value=p.BasePrice!=null?p.BasePrice:''; if(fReady) fReady.checked=!!p.ReadyToShip; if(fWidth) fWidth.value=p.WidthMM||0; if(fHeight) fHeight.value=p.HeightMM||0; if(fDepth) fDepth.value=p.DepthMM||0; if(btnDel) btnDel.style.display=''; setModeEdit(true);
    renderGallery((p && p.Images) || []);
    scrollToForm();
  }
  function clear(){ if(form) form.reset(); if(fSlug) fSlug.value=''; if(btnDel) btnDel.style.display='none'; if(btnManage) btnManage.style.display='none'; if(fWidth) fWidth.value=''; if(fHeight) fHeight.value=''; if(fDepth) fDepth.value=''; if(imagesInput) imagesInput.value=''; if(preview) preview.innerHTML=''; if(dzCount) dzCount.textContent='0 archivos'; setModeEdit(false); selectedIDs.clear(); updateSelectionUI(); renderGallery([]); }

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
        if(res.ok){ selectedIDs.delete(im.ID); updateSelectionUI(); card.remove(); } else { alert('Error eliminando imagen'); }
      });
      del.addEventListener('click', e=>e.stopPropagation());
      card.appendChild(img); card.appendChild(del); card.appendChild(sel); g.appendChild(card);
    });
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

  form.addEventListener('submit', async e=>{
    e.preventDefault();
    const slug=(fSlug&&fSlug.value.trim())||'';
    const payload={ name:(fName&&fName.value.trim())||'', category:(fCat&&fCat.value.trim())||'', short_desc:(fDesc&&fDesc.value)||'', base_price:parseFloat((fPrice&&fPrice.value)||'0'), ready_to_ship:!!(fReady&&fReady.checked), width_mm:parseFloat((fWidth&&fWidth.value)||'0'), height_mm:parseFloat((fHeight&&fHeight.value)||'0'), depth_mm:parseFloat((fDepth&&fDepth.value)||'0') };
    if(!payload.name){ alert('Nombre requerido'); return; }
    if(payload.base_price<0){ alert('Precio inválido'); return; }
    let method='POST', url='/api/products'; if(slug){ method='PUT'; url='/api/products/'+encodeURIComponent(slug); }
    const res=await fetch(url,{method, headers:Object.assign({'Content-Type':'application/json'}, token? {Authorization:'Bearer '+token}:{}) , body:JSON.stringify(payload)});
    if(!res.ok){ alert('Error guardando'); return; }
    const prod=await res.json(); const finalSlug=(prod&&prod.Slug)||slug;
    if(imagesInput && imagesInput.files && imagesInput.files.length>0){
      const fd=new FormData(); fd.append('existing_slug', finalSlug);
      for(const f of imagesInput.files){ fd.append('images', f); }
      const upRes=await fetch('/api/products/upload',{method:'POST', headers: token? {Authorization:'Bearer '+token}:{}, body:fd});
      if(!upRes.ok){ alert('Producto guardado, pero error subiendo imágenes'); location.reload(); return; }
    }
    location.reload();
  });

  if(btnReset) btnReset.addEventListener('click', clear);
  if(btnDel) btnDel.addEventListener('click', async ()=>{ const slug=(fSlug&&fSlug.value.trim())||''; if(!slug) return; if(!confirm('Eliminar producto y sus imágenes?')) return; const res=await fetch('/api/products/'+encodeURIComponent(slug),{method:'DELETE', headers: token? {Authorization:'Bearer '+token}:{}}); if(res.ok){ clear(); location.reload(); } else { alert('Error'); } });

  function refreshPreview(){ if(!preview||!imagesInput||!dzCount) return; preview.innerHTML=''; const files=Array.from(imagesInput.files||[]); dzCount.textContent=files.length+(files.length===1?' archivo':' archivos'); files.slice(0,6).forEach(f=>{ const r=new FileReader(); r.onload=ev=>{ const img=document.createElement('img'); img.src=ev.target.result; img.alt=f.name; img.style.width='52px'; img.style.height='52px'; img.style.objectFit='cover'; img.style.borderRadius='10px'; img.style.border='1px solid #223140'; preview.appendChild(img); }; r.readAsDataURL(f); }); }
  if(imagesInput) imagesInput.addEventListener('change', refreshPreview);
  if(dropZone){ ['dragenter','dragover'].forEach(ev=>dropZone.addEventListener(ev,e=>{e.preventDefault(); dropZone.classList.add('drag');})); ['dragleave','drop'].forEach(ev=>dropZone.addEventListener(ev,e=>{e.preventDefault(); dropZone.classList.remove('drag');})); dropZone.addEventListener('drop', e=>{ const files=[...e.dataTransfer.files].filter(f=>f.type.startsWith('image/')); if(files.length){ const dt=new DataTransfer(); files.forEach(f=>dt.items.add(f)); if(imagesInput) imagesInput.files=dt.files; refreshPreview(); }}); }

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
  if(mgrDelBtn){ mgrDelBtn.addEventListener('click', async (e)=>{ e.preventDefault(); if(mgrSelected.size===0) return; if(!token){ alert('Sesión no válida'); return; } if(!confirm('Eliminar imágenes seleccionadas?')) return; const ids=[...mgrSelected]; await Promise.all(ids.map(id=>fetch('/api/product_images/'+encodeURIComponent(id),{method:'DELETE', headers:{Authorization:'Bearer '+token}}))); ids.forEach(id=>{ const el=grid && grid.querySelector('.img-card[data-id="'+CSS.escape(id)+'"]'); if(el) el.remove(); }); mgrSelected.clear(); mgrUpdateSel(); }); }
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
          ['Material', `$${out.precio_material.toFixed(2)}`],
          ['Luz', `$${out.precio_luz.toFixed(2)}`],
          ['Error', `$${out.margen_de_error.toFixed(2)}`],
          ['Subtotal s/ins.', `$${out.subtotal_sin_insumos.toFixed(2)}`],
          ['Total s/ins.', `$${out.total_sin_insumos.toFixed(2)}`],
          ['Insumos', `$${out.insumos.toFixed(2)}`],
          ['Total a cobrar', `$${out.total_a_cobrar.toFixed(2)}`],
          ['Precio ML', `$${out.precio_mercadolibre.toFixed(2)}`],
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