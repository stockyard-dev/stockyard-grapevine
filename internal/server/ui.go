package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Grapevine</title>
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:ital,wght@0,400;0,700;1,400&family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c94444;--orange:#d4843a;--blue:#5b8dd9;--mono:'JetBrains Mono',monospace;--serif:'Libre Baskerville',serif}
*{margin:0;padding:0;box-sizing:border-box}
body{background:var(--bg);color:var(--cream);font-family:var(--mono);line-height:1.5;font-size:13px}
.hdr{padding:.8rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center;gap:1rem;flex-wrap:wrap}
.hdr h1{font-size:.9rem;letter-spacing:2px}
.hdr h1 span{color:var(--rust)}
.main{padding:1.2rem 1.5rem;max-width:980px;margin:0 auto}
.stats{display:grid;grid-template-columns:repeat(4,1fr);gap:.5rem;margin-bottom:1rem}
.st{background:var(--bg2);border:1px solid var(--bg3);padding:.7rem;text-align:center}
.st-v{font-size:1.2rem;font-weight:700;color:var(--gold)}
.st-v.green{color:var(--green)}
.st-l{font-size:.5rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-top:.2rem}
.toolbar{display:flex;gap:.5rem;margin-bottom:1rem;flex-wrap:wrap;align-items:center}
.search{flex:1;min-width:180px;padding:.4rem .6rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.search:focus{outline:none;border-color:var(--leather)}
.filter-sel{padding:.4rem .5rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.65rem}

.list{display:flex;flex-direction:column;gap:.6rem}
.art{background:var(--bg2);border:1px solid var(--bg3);padding:1rem 1.1rem;cursor:pointer;display:flex;flex-direction:column;gap:.4rem;transition:border-color .15s}
.art:hover{border-color:var(--leather)}
.art.draft{opacity:.65}
.art.archived{opacity:.5}
.art-hdr{display:flex;justify-content:space-between;align-items:flex-start;gap:.5rem}
.art-title{font-family:var(--serif);font-size:1rem;font-weight:700;color:var(--cream);flex:1}
.art-body{font-size:.7rem;color:var(--cd);line-height:1.5;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden}
.art-meta{display:flex;gap:.5rem;flex-wrap:wrap;align-items:center;font-size:.55rem;color:var(--cm)}
.badge{font-size:.5rem;padding:.12rem .35rem;text-transform:uppercase;letter-spacing:1px;border:1px solid var(--bg3);color:var(--cm);font-weight:700}
.badge.published{border-color:var(--green);color:var(--green)}
.badge.draft{border-color:var(--orange);color:var(--orange)}
.badge.archived{border-color:var(--cm);color:var(--cm)}
.badge.cat{border-color:var(--leather);color:var(--leather)}
.tag{font-size:.5rem;padding:.05rem .3rem;background:var(--bg3);color:var(--cd);font-family:var(--mono)}
.helpful-row{display:flex;gap:.5rem;align-items:center;margin-top:.3rem}
.helpful-bar{flex:1;height:6px;background:var(--bg3);position:relative;overflow:hidden}
.helpful-bar-fill{position:absolute;top:0;left:0;bottom:0;background:var(--green);transition:width .2s}
.helpful-counts{font-family:var(--mono);font-size:.55rem;color:var(--cm);min-width:80px;text-align:right}
.helpful-counts strong.up{color:var(--green)}
.helpful-counts strong.down{color:var(--red)}
.helpful-actions{display:flex;gap:.3rem;margin-top:.3rem}
.h-btn{font-family:var(--mono);font-size:.55rem;padding:.2rem .4rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd);transition:.15s}
.h-btn:hover.up{border-color:var(--green);color:var(--green)}
.h-btn:hover.down{border-color:var(--red);color:var(--red)}

.btn{font-family:var(--mono);font-size:.6rem;padding:.3rem .55rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd);transition:.15s}
.btn:hover{border-color:var(--leather);color:var(--cream)}
.btn-p{background:var(--rust);border-color:var(--rust);color:#fff}
.btn-p:hover{opacity:.85;color:#fff}
.btn-sm{font-size:.55rem;padding:.2rem .4rem}
.btn-del{color:var(--red);border-color:#3a1a1a}
.btn-del:hover{border-color:var(--red);color:var(--red)}

.modal-bg{display:none;position:fixed;inset:0;background:rgba(0,0,0,.65);z-index:100;align-items:center;justify-content:center}
.modal-bg.open{display:flex}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:600px;max-width:92vw;max-height:90vh;overflow-y:auto}
.modal h2{font-size:.8rem;margin-bottom:1rem;color:var(--rust);letter-spacing:1px}
.fr{margin-bottom:.6rem}
.fr label{display:block;font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.2rem}
.fr input,.fr select,.fr textarea{width:100%;padding:.4rem .5rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.fr textarea{font-family:var(--serif);font-size:.85rem;line-height:1.5}
.fr input:focus,.fr select:focus,.fr textarea:focus{outline:none;border-color:var(--leather)}
.row2{display:grid;grid-template-columns:1fr 1fr;gap:.5rem}
.fr-section{margin-top:1rem;padding-top:.8rem;border-top:1px solid var(--bg3)}
.fr-section-label{font-size:.55rem;color:var(--rust);text-transform:uppercase;letter-spacing:1px;margin-bottom:.5rem}
.acts{display:flex;gap:.4rem;justify-content:flex-end;margin-top:1rem}
.acts .btn-del{margin-right:auto}
.empty{text-align:center;padding:3rem;color:var(--cm);font-style:italic;font-size:.85rem}
@media(max-width:600px){.stats{grid-template-columns:repeat(2,1fr)}}
</style>
</head>
<body>

<div class="hdr">
<h1 id="dash-title"><span>&#9670;</span> GRAPEVINE</h1>
<button class="btn btn-p" onclick="openNew()">+ New Article</button>
</div>

<div class="main">
<div class="stats" id="stats"></div>
<div class="toolbar">
<input class="search" id="search" placeholder="Search articles..." oninput="debouncedRender()">
<select class="filter-sel" id="status-filter" onchange="render()">
<option value="">All Statuses</option>
<option value="published">Published</option>
<option value="draft">Draft</option>
<option value="archived">Archived</option>
</select>
<select class="filter-sel" id="category-filter" onchange="render()">
<option value="">All Categories</option>
</select>
</div>
<div id="list" class="list"></div>
</div>

<div class="modal-bg" id="mbg" onclick="if(event.target===this)closeModal()">
<div class="modal" id="mdl"></div>
</div>

<script>
var A='/api';
var RESOURCE='articles';

var fields=[
{name:'title',label:'Title',type:'text',required:true},
{name:'body',label:'Body (Markdown)',type:'textarea'},
{name:'category',label:'Category',type:'select_or_text',options:[]},
{name:'tags',label:'Tags',type:'text',placeholder:'comma separated'},
{name:'slug',label:'Slug (auto if empty)',type:'text'},
{name:'status',label:'Status',type:'select',options:['published','draft','archived']}
];

var articles=[],artExtras={},editId=null,searchTimer=null;

function fmtDate(s){
if(!s)return'';
try{
var d=new Date(s);
if(isNaN(d.getTime()))return s;
return d.toLocaleDateString('en-US',{month:'short',day:'numeric',year:'numeric'});
}catch(e){return s}
}

function fieldByName(n){for(var i=0;i<fields.length;i++)if(fields[i].name===n)return fields[i];return null}

function debouncedRender(){
clearTimeout(searchTimer);
searchTimer=setTimeout(render,200);
}

async function load(){
try{
var resps=await Promise.all([
fetch(A+'/articles').then(function(r){return r.json()}),
fetch(A+'/stats').then(function(r){return r.json()})
]);
articles=resps[0].articles||[];
renderStats(resps[1]||{});

try{
var ex=await fetch(A+'/extras/'+RESOURCE).then(function(r){return r.json()});
artExtras=ex||{};
articles.forEach(function(a){
var x=artExtras[a.id];
if(!x)return;
Object.keys(x).forEach(function(k){if(a[k]===undefined)a[k]=x[k]});
});
}catch(e){artExtras={}}

populateCategoryFilter();
}catch(e){
console.error('load failed',e);
articles=[];
}
render();
}

function populateCategoryFilter(){
var sel=document.getElementById('category-filter');
if(!sel)return;
var current=sel.value;
var seen={};var cats=[];
articles.forEach(function(a){if(a.category&&!seen[a.category]){seen[a.category]=true;cats.push(a.category)}});
cats.sort();
sel.innerHTML='<option value="">All Categories</option>'+cats.map(function(c){return'<option value="'+esc(c)+'"'+(c===current?' selected':'')+'>'+esc(c)+'</option>'}).join('');
}

function renderStats(s){
var total=s.total||0;
var helpful=s.total_helpful||0;
var unhelpful=s.total_unhelpful||0;
var byStatus=s.by_status||{};
var pub=byStatus.published||0;
document.getElementById('stats').innerHTML=
'<div class="st"><div class="st-v">'+total+'</div><div class="st-l">Articles</div></div>'+
'<div class="st"><div class="st-v green">'+pub+'</div><div class="st-l">Published</div></div>'+
'<div class="st"><div class="st-v">'+helpful+'</div><div class="st-l">Helpful Votes</div></div>'+
'<div class="st"><div class="st-v">'+unhelpful+'</div><div class="st-l">Not Helpful</div></div>';
}

function render(){
var q=(document.getElementById('search').value||'').toLowerCase();
var sf=document.getElementById('status-filter').value;
var cf=document.getElementById('category-filter').value;

var f=articles.slice();
if(q)f=f.filter(function(a){
return(a.title||'').toLowerCase().includes(q)||
(a.body||'').toLowerCase().includes(q)||
(a.tags||'').toLowerCase().includes(q);
});
if(sf)f=f.filter(function(a){return a.status===sf});
if(cf)f=f.filter(function(a){return a.category===cf});

if(!f.length){
var msg=window._emptyMsg||'No articles yet.';
document.getElementById('list').innerHTML='<div class="empty">'+esc(msg)+'</div>';
return;
}

var h='';
f.forEach(function(a){h+=articleHTML(a)});
document.getElementById('list').innerHTML=h;
}

function articleHTML(a){
var helpful=a.helpful||0;
var unhelpful=a.not_helpful||0;
var totalVotes=helpful+unhelpful;
var pct=totalVotes>0?Math.round((helpful/totalVotes)*100):0;

var cls='art '+(a.status||'published');

var h='<div class="'+cls+'">';
h+='<div class="art-hdr" onclick="openEdit(\''+esc(a.id)+'\')">';
h+='<div class="art-title">'+esc(a.title)+'</div>';
h+='</div>';

if(a.body)h+='<div class="art-body" onclick="openEdit(\''+esc(a.id)+'\')">'+esc(a.body)+'</div>';

h+='<div class="art-meta">';
if(a.status)h+='<span class="badge '+esc(a.status)+'">'+esc(a.status)+'</span>';
if(a.category)h+='<span class="badge cat">'+esc(a.category)+'</span>';
if(a.slug)h+='<span>/'+esc(a.slug)+'</span>';
h+='<span>'+esc(fmtDate(a.created_at))+'</span>';
h+='</div>';

if(a.tags){
var tagList=String(a.tags).split(',').map(function(t){return t.trim()}).filter(function(t){return t});
if(tagList.length){
h+='<div style="display:flex;gap:.3rem;flex-wrap:wrap">';
tagList.forEach(function(t){h+='<span class="tag">#'+esc(t)+'</span>'});
h+='</div>';
}
}

// Helpful bar
h+='<div class="helpful-row">';
h+='<div class="helpful-bar"><div class="helpful-bar-fill" style="width:'+pct+'%"></div></div>';
h+='<div class="helpful-counts"><strong class="up">'+helpful+'</strong> / <strong class="down">'+unhelpful+'</strong></div>';
h+='</div>';

h+='<div class="helpful-actions">';
h+='<button class="h-btn up" onclick="markHelpful(\''+esc(a.id)+'\',event)">&#128077; Helpful</button>';
h+='<button class="h-btn down" onclick="markNotHelpful(\''+esc(a.id)+'\',event)">&#128078; Not Helpful</button>';
h+='</div>';

h+='</div>';
return h;
}

async function markHelpful(id,ev){
ev.stopPropagation();
try{
await fetch(A+'/articles/'+id+'/helpful',{method:'POST'});
load();
}catch(e){}
}

async function markNotHelpful(id,ev){
ev.stopPropagation();
try{
await fetch(A+'/articles/'+id+'/not-helpful',{method:'POST'});
load();
}catch(e){}
}

// ─── Modal ────────────────────────────────────────────────────────

function fieldHTML(f,value){
var v=value;
if(v===undefined||v===null)v='';
var req=f.required?' *':'';
var ph=f.placeholder?(' placeholder="'+esc(f.placeholder)+'"'):'';
var h='<div class="fr"><label>'+esc(f.label)+req+'</label>';

if(f.type==='select'){
h+='<select id="f-'+f.name+'">';
if(!f.required)h+='<option value="">Select...</option>';
(f.options||[]).forEach(function(o){
var sel=(String(v)===String(o))?' selected':'';
h+='<option value="'+esc(String(o))+'"'+sel+'>'+esc(String(o))+'</option>';
});
h+='</select>';
}else if(f.type==='select_or_text'){
h+='<input list="dl-'+f.name+'" type="text" id="f-'+f.name+'" value="'+esc(String(v))+'"'+ph+'>';
h+='<datalist id="dl-'+f.name+'">';
var opts=(f.options||[]).slice();
articles.forEach(function(a){if(a.category&&opts.indexOf(a.category)===-1)opts.push(a.category)});
opts.forEach(function(o){h+='<option value="'+esc(String(o))+'">'});
h+='</datalist>';
}else if(f.type==='textarea'){
h+='<textarea id="f-'+f.name+'" rows="10"'+ph+'>'+esc(String(v))+'</textarea>';
}else{
h+='<input type="text" id="f-'+f.name+'" value="'+esc(String(v))+'"'+ph+'>';
}
h+='</div>';
return h;
}

function formHTML(art){
var a=art||{};
var isEdit=!!art;
var h='<h2>'+(isEdit?'EDIT ARTICLE':'NEW ARTICLE')+'</h2>';

h+=fieldHTML(fieldByName('title'),a.title);
h+=fieldHTML(fieldByName('body'),a.body);
h+='<div class="row2">'+fieldHTML(fieldByName('category'),a.category)+fieldHTML(fieldByName('status'),a.status||'published')+'</div>';
h+='<div class="row2">'+fieldHTML(fieldByName('tags'),a.tags)+fieldHTML(fieldByName('slug'),a.slug)+'</div>';

var customFields=fields.filter(function(f){return f.isCustom});
if(customFields.length){
var label=window._customSectionLabel||'Additional Details';
h+='<div class="fr-section"><div class="fr-section-label">'+esc(label)+'</div>';
customFields.forEach(function(f){h+=fieldHTML(f,a[f.name])});
h+='</div>';
}

h+='<div class="acts">';
if(isEdit)h+='<button class="btn btn-del" onclick="delItem()">Delete</button>';
h+='<button class="btn" onclick="closeModal()">Cancel</button>';
h+='<button class="btn btn-p" onclick="submit()">'+(isEdit?'Save':'Publish')+'</button>';
h+='</div>';
return h;
}

function openNew(){
editId=null;
document.getElementById('mdl').innerHTML=formHTML();
document.getElementById('mbg').classList.add('open');
var t=document.getElementById('f-title');if(t)t.focus();
}

function openEdit(id){
var a=null;
for(var i=0;i<articles.length;i++){if(articles[i].id===id){a=articles[i];break}}
if(!a)return;
editId=id;
document.getElementById('mdl').innerHTML=formHTML(a);
document.getElementById('mbg').classList.add('open');
}

function closeModal(){
document.getElementById('mbg').classList.remove('open');
editId=null;
}

async function submit(){
var titleEl=document.getElementById('f-title');
if(!titleEl||!titleEl.value.trim()){alert('Title is required');return}

var body={};
var extras={};
fields.forEach(function(f){
var el=document.getElementById('f-'+f.name);
if(!el)return;
var val=el.value.trim();
if(f.isCustom)extras[f.name]=val;
else body[f.name]=val;
});

var savedId=editId;
try{
if(editId){
var r1=await fetch(A+'/articles/'+editId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
if(!r1.ok){var e1=await r1.json().catch(function(){return{}});alert(e1.error||'Save failed');return}
}else{
var r2=await fetch(A+'/articles',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
if(!r2.ok){var e2=await r2.json().catch(function(){return{}});alert(e2.error||'Publish failed');return}
var created=await r2.json();
savedId=created.id;
}
if(savedId&&Object.keys(extras).length){
await fetch(A+'/extras/'+RESOURCE+'/'+savedId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(extras)}).catch(function(){});
}
}catch(e){alert('Network error: '+e.message);return}
closeModal();
load();
}

async function delItem(){
if(!editId)return;
if(!confirm('Delete this article?'))return;
await fetch(A+'/articles/'+editId,{method:'DELETE'});
closeModal();
load();
}

function esc(s){
if(s===undefined||s===null)return'';
var d=document.createElement('div');
d.textContent=String(s);
return d.innerHTML;
}

document.addEventListener('keydown',function(e){if(e.key==='Escape')closeModal()});

(function loadPersonalization(){
fetch('/api/config').then(function(r){return r.json()}).then(function(cfg){
if(!cfg||typeof cfg!=='object')return;

if(cfg.dashboard_title){
var h1=document.getElementById('dash-title');
if(h1)h1.innerHTML='<span>&#9670;</span> '+esc(cfg.dashboard_title);
document.title=cfg.dashboard_title;
}

if(cfg.empty_state_message)window._emptyMsg=cfg.empty_state_message;
if(cfg.primary_label)window._customSectionLabel=cfg.primary_label+' Details';

if(Array.isArray(cfg.categories)){
var catField=fieldByName('category');
if(catField)catField.options=cfg.categories;
}

if(Array.isArray(cfg.custom_fields)){
cfg.custom_fields.forEach(function(cf){
if(!cf||!cf.name||!cf.label)return;
if(fieldByName(cf.name))return;
fields.push({
name:cf.name,
label:cf.label,
type:cf.type||'text',
options:cf.options||[],
isCustom:true
});
});
}
}).catch(function(){
}).finally(function(){
load();
});
})();
</script>
</body>
</html>`
