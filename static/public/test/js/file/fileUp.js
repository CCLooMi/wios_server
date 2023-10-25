//created by chenxianjun at 2019.01.18
(function (global, factory) {
    if(typeof exports === 'object' && typeof module !=='undefined'){
        factory(global,exports);
    }else if(typeof define==='function'&&define.amd){
        define(['exports'],function (exports) {
            factory(global,exports);
        });
    }else {
        global=global||self;
        factory(global,global);
    }
}(this, function (global,exports) {
    function ce(name) {
        return document.createElement(name);
    }
    function cdf() {
        return document.createDocumentFragment();
    }
    function attacheEvent(ele) {
        if(!ele||!ele.addEventListener){
            return{on:()=>this,destroy:()=>this,watch:()=>this};
        }
        var eventDispose=[];
        return {
            'on':function (type,func) {
                ele.addEventListener(type,func);
                eventDispose.push(function(){
                    ele.removeEventListener(type,func);
                });
                return this;
            },
            'getDispose':function (f) {
                var dsp=function () {
                    for(var i=0;i<eventDispose.length;i++){
                        eventDispose[i]();
                        delete eventDispose[i];
                    }
                    eventDispose=null;
                };
                if(typeof f == 'function'){
                    f(dsp);
                }
                return dsp;
            }
        };
    }
    function httpGet(url) {
        return new Promise(function (resolve, reject) {
            try{
                var xhr=new XMLHttpRequest();
                xhr.responseType='blob';
                var dispose=attacheEvent(xhr)
                    .on('error',reject)
                    .on('abort',reject)
                    .on('timeout',reject)
                    .on('loadend',function (e) {
                        dispose();
                    })
                    .getDispose();
                xhr.onreadystatechange = function () {
                    if (xhr.readyState === 4 && xhr.status === 200) {
                        resolve(xhr);
                    }else if(xhr.status&&xhr.status!=200){
                        reject(new Error(xhr.status+':['+xhr.responseURL+']:'+xhr.statusText));
                    }
                };
                xhr.open("GET", url, true);
                xhr.send();
            }catch (e) {
                reject(e);
            }
        });
    }
    function hexToBytes(hex){
        let bytes=[];
        //填充到指定长度，否则当hex末尾为00时，生成bytes长度将不够
        bytes.push.apply(bytes,new Uint8Array((hex.length+1)/2));
        for(let i=0,j=0;i<hex.length;i++) {
            switch (hex.charAt(i)) {
                case '1':bytes[j]|=0x10>>((i&1)<<2);break;
                case '2':bytes[j]|=0x20>>((i&1)<<2);break;
                case '3':bytes[j]|=0x30>>((i&1)<<2);break;
                case '4':bytes[j]|=0x40>>((i&1)<<2);break;
                case '5':bytes[j]|=0x50>>((i&1)<<2);break;
                case '6':bytes[j]|=0x60>>((i&1)<<2);break;
                case '7':bytes[j]|=0x70>>((i&1)<<2);break;
                case '8':bytes[j]|=0x80>>((i&1)<<2);break;
                case '9':bytes[j]|=0x90>>((i&1)<<2);break;
                case 'a':case 'A':bytes[j]|=0xA0>>((i&1)<<2);break;
                case 'b':case 'B':bytes[j]|=0xB0>>((i&1)<<2);break;
                case 'c':case 'C':bytes[j]|=0xC0>>((i&1)<<2);break;
                case 'd':case 'D':bytes[j]|=0xD0>>((i&1)<<2);break;
                case 'e':case 'E':bytes[j]|=0xE0>>((i&1)<<2);break;
                case 'f':case 'F':bytes[j]|=0xF0>>((i&1)<<2);break;
                default:break;
            }
            j+=i&1;
        }
        return new Uint8Array(bytes);
    }

    //for watchInDomTree
    function clearObjProperties(...objs) {
        if(Array.isArray(objs[0])){
            objs=objs[0];
        }
        for(var i=0;i<objs.length;i++){
            var obj=objs[i];
            var ks=Object.keys(obj);
            for(var ii=0;ii<ks.length;ii++){
                delete obj[ks[ii]];
            }
        }
    }
    const csm = [];
    (function () {
        var cs = "0123456789abcdef";
        var n = 0;
        for (var i = 0; i < 16; i++) {
            for (var j = 0; j < 16; j++, n++) {
                csm[n] = cs[i] + cs[j];
            }
        }
    })();
    function bytesToHex(a) {
        var s = '';
        for (var i = 0; i < a.length; i++) {
            s += csm[a[i] & 0xff];
        }
        return s;
    }
    function uuid(len) {
        len = len>>1||16;
        if(len<1){len=16;}
        const bid = new ArrayBuffer(len);
        const dv = new DataView(bid);
        for(let i=0;i<len;i++){
            dv.setUint8(i,Math.random()*255|0);
        }
        return bytesToHex(new Uint8Array(bid));
    }
    window.uuid=uuid;
    if(!Object.prototype._uuid){
        Object.defineProperty(Object.prototype,'_uuid',{
            enumerable:false,
            get:function () {
                if(this.__uuid===undefined){
                    Object.defineProperty(this,'__uuid',{
                        value:uuid(32),
                        enumerable:false,
                        writable:false
                    });
                }
                return this.__uuid;
            }
        });
    }
    function isInPage(node) {
        return node.isConnected||document.contains(node);
    }
    const watchSet={};
    function watchInDomTree(ele,callback) {
        if(!ele||!(ele instanceof Node||ele instanceof Attr)||(typeof callback!='function')){
            return;
        }
        if(!isInPage(ele)){
            console.warn([ele,'not in page.']);
        }
        var _wf=watchSet[ele._uuid];
        if(_wf){
            var cbs=_wf.callbacks;
            cbs.push(callback);
            return function () {
                cbs.splice(cbs.indexOf(callback),1);
                if(!cbs.length){
                    _wf.release=true;
                }
            };
        }
        var wf=function (){
            if(!isInPage(ele)){
                for(var i=0;i<wf.callbacks.length;i++){
                    try{
                        wf.callbacks[i](ele);
                        clearObjProperties(ele);
                    }catch (e) {}
                }
                return true;
            }
            if(wf.release){
                return true;
            }
            return false;
        };
        wf.callbacks=[callback];
        watchSet[ele._uuid]=wf;
        if(Object.keys(watchSet).length==1){
            startWatch();
        }
        return function (){
            var cbs=wf.callbacks;
            cbs.splice(cbs.indexOf(callback),1);
            if(!cbs.length){
                wf.release=true;
            }
        };
    }
    function startWatch() {
        startWatch.to&&clearTimeout(startWatch.to);
        if(Object.keys(watchSet).length){
            for(var p in watchSet){
                if(watchSet[p]()){
                    delete watchSet[p];
                }
            }
        }
        if(Object.keys(watchSet).length){
            startWatch.to=setTimeout(startWatch,200);
        }
    }
    //end

    var initWK;
    function initWKBlob(worker,...deps) {
        if(!initWK){
            initWK=new Promise(function (resolve, reject) {
                const ps=[];
                if(worker instanceof Blob){
                    ps.push(worker);
                }else{
                    ps.push(httpGet(worker));
                }
                for(let i=0;i<deps.length;i++){
                    ps.push(httpGet(deps[i]));
                }
                Promise.all(ps)
                    .then(a=>a.map(ia=>ia.response))
                    .then(blobs=>{
                        const b1=blobs.shift();
                        const scripts=blobs.map(b=>`"${URL.createObjectURL(b)}"`).join(',');
                        const wksBlob=new Blob(
                            [`importScripts(${scripts});`,b1],
                            {type:'application/javascript'});
                        resolve(wksBlob);
                    },reject)
            });
        }
        return initWK;
    }
    function readFile(reader, file, block) {
        return new Promise(function (resolve, reject) {
            reader.onload = function (e) {
                resolve({'msg': e.target.result, 'block': block});
            };
            reader.onerror = reject;
            reader.readAsArrayBuffer(file.slice(block.start, block.end));
        })
    }
    function readFileByBlock(reader,file, block) {
        return new Promise(function (resolve, reject) {
            reader.onload = function (e) {
                resolve(e.target.result);
            }
            reader.onerror=reject;
            reader.readAsArrayBuffer(file.slice(block.start, block.end));
        })
    }
    function newFileInput(multiple,accept,capture) {
        var finput = ce('input');
        finput.type = "file";
        finput.multiple=multiple;
        finput.accept=accept;
        finput.capture=capture;
        return finput;
    }
    function getFiles(e) {
        e.stopPropagation(),e.preventDefault();
        return [...(e.dataTransfer?.files || e.target.files || [])];
    }
    const BUF_SIZE = 1024*1024;
    function doHash(algo, file, worker) {
        return new Promise(function (resolve, reject) {
            var reader = new FileReader();
            var dsp = attacheEvent(worker)
                .on('message', e => {
                    var d = e.data;
                    if (d.error) {
                        dsp();
                        reject(d.error);
                        return;
                    }
                    var block = d.block;
                    if (file.progress instanceof Function) {
                        var time = Date.now() - block.time;
                        var speed = parseInt(block.end / time / 1000, 10) + 'MB/s';
                        file.progress({
                            type:'hash',
                            loaded: block.end,
                            total: file.size,
                            blockSize: block.end - block.start,
                            progress: block.end / file.size,
                            result: d.result,
                            speed: speed
                        });
                    }
                    if (d.result) {
                        dsp();
                        resolve(d.result);
                        return;
                    }
                    if (block.end != file.size) {
                        block.start += BUF_SIZE;
                        block.end += BUF_SIZE;
                        if (block.end > file.size) {
                            block.end = file.size;
                        }
                        readFile(reader, file, block)
                            .then(r => worker.postMessage(r),reject);
                    }
                })
                .getDispose();
            readFile(reader, file, {
                time: Date.now(),
                fileSize: file.size,
                start: 0,
                end: (BUF_SIZE > file.size ? file.size : BUF_SIZE)
            }).then(r => {
                r.algo = algo;
                worker.postMessage(r);
            },reject);
        });
    }
    function doUpload(file,channel,callback) {
        var reader = new FileReader();
        function pushData(arrayBuf,cmd,channel,callback) {
            const bid = hexToBytes(file.id);
            const dataLen = 4+bid.length+arrayBuf.byteLength;
            const dataAB = new ArrayBuffer(dataLen);
            const dataView = new DataView(dataAB);
            const dataUA = new Uint8Array(dataAB);
            dataView.setUint32(0,bid.length);
            dataUA.set(bid,4);
            dataUA.set(new Uint8Array(arrayBuf),4+bid.length);
            channel.pushFileData(dataAB)
                .receive(function (cmd) {
                    fileProgress(file,cmd)
                    if(cmd.complete==1){
                        callback(file,channel);
                    }else{
                        readFileByBlock(reader,file,cmd)
                            .then(ab=>pushData(ab,cmd,channel,callback),e=>callback(file,channel,e));
                    }
                });
        }
        function fileProgress(file,cmd) {
            if (file.progress instanceof Function) {
                file.progress({
                    type:'upload',
                    loaded: cmd.complete*100,
                    total: 100,
                    progress: cmd.complete
                });
            }
        }
        channel.pushFileInfo({id:file.id,size:file.size,name:file.name})
            .receive(function (cmd) {
                fileProgress(file,cmd)
                if(cmd.complete==1){
                    callback(file,channel);
                }else{
                    readFileByBlock(reader,file,cmd)
                        .then(ab=>pushData(ab,cmd,channel,callback),e=>callback(file,channel,e));
                }
            });
    }
    function Channel(url){
        this.url=url||
            this.url||
            location.origin.replace('http', 'ws')+'/fileUp';
        this.callbacks=[];
        function callback() {
            let c;
            while((c=this.callbacks.shift())){
                c.apply(this,arguments);
            }
        }
        function newWebSocket() {
            this.socket=new WebSocket(this.url);
            this.dispose=attacheEvent(this.socket)
                .on('open',function () {
                    console.log(`Connected to [ ${this.url} ]`);
                })
                .on('close',()=> {
                    this.dispose();
                    console.log(`Disconnected from [ ${this.url} ]`);
                    setTimeout( ()=> {
                        console.log(`Reconnecting to [ ${this.url} ]`);
                        newWebSocket.call(this);
                    },500);
                })
                .on('message',e=>{
                    e.data.arrayBuffer().then( data=> {
                        let dv = new DataView(data);
                        const idLen = Number(dv.getInt32(0));
                        const id = new Uint8Array(data,4,idLen);
                        if(data.byteLength>idLen){
                            const start = Number(dv.getBigInt64(idLen+4));
                            const end = Number(dv.getBigInt64(idLen+4+8));
                            const uploaded = Number(dv.getBigInt64(idLen+4+16));
                            const total = Number(dv.getBigInt64(idLen+4+24));
                            const idx = Number(dv.getBigInt64(idLen+4+32));
                            callback.call(this,{
                                id:bytesToHex(id),
                                start:start,end:end,
                                uploaded:uploaded,total:total,
                                complete:uploaded/total,idx:idx,
                            });
                            return;
                        }
                        callback.call(this,{id:bytesToHex(id),complete:1});
                    },console.error);
                })
                .getDispose();
        }
        newWebSocket.call(this);
    }
    Channel.prototype={
        pushFileInfo:function (infoObj) {
            const $this = this;
            this.socket.send(JSON.stringify(infoObj));
            return {
                receive:function (func) {
                    if(func instanceof Function){
                        $this.callbacks.push(func);
                    }
                }
            }
        },
        pushFileData:function (data) {
            const $this = this;
            this.socket.send(data);
            return {
                receive:function (func) {
                    if(func instanceof Function){
                        $this.callbacks.push(func);
                    }
                }
            }
        }
    }
    function FileUp(ele,option) {
        option=option||{};
        this.filesToHash=[];
        this.filesToUpload=[];
        this.workerCount=0;
        this.maxWorkers=option.maxWorkers||3;
        this.fileSelect= e=> {
            var files = getFiles(e);
            if(option.fileSelect instanceof Function){
                option.fileSelect(files);
            }
            this.filesToHash.push(...files);
            this.startHash('SHA256',null,option.onComplete);
        }
        this.channels=[new Channel(option.uploadUrl)];
        initWKBlob(option.worker,...option.deps||[])
            .then(wkBlob=>{
                this.wkBlob=wkBlob;
                this.workerUrl=URL.createObjectURL(wkBlob);
                this.finput=newFileInput(option.multiple,option.accept,option.capture);
                this.finput.onchange=this.fileSelect;
                attacheEvent(ele)
                    .on('dragover', e => (e.preventDefault(), e.stopPropagation()))
                    .on('drop', this.fileSelect)
                    .on('click', e => this.finput.click())
                    .getDispose(dsp=>watchInDomTree(ele,dsp));
            });
    }
    FileUp.prototype={
        startHash:function (algo, worker, callback) {
            var $this=this;
            while (worker || this.workerCount < this.maxWorkers) {
                var file = this.filesToHash.shift();
                if (!file) {
                    if (worker) {
                        this.workerCount--;
                        worker.terminate();
                    }
                    break;
                }
                var _worker = worker || new Worker(this.workerUrl);
                if (!worker) {
                    this.workerCount++;
                }
                file.worker = _worker;
                (function (file,worker) {
                    doHash(algo, file, worker).then(hash=>{
                        file.id=hash;
                        $this.filesToUpload.push(file);
                        $this.startUpload(callback);
                        //hashNextFile
                        $this.startHash(algo,worker);
                    });
                })(file,_worker);
                if (worker) {
                    break;
                }
            }
        },
        startUpload:function (callback) {
            var channel=this.channels.shift();
            if(channel){
                var file = this.filesToUpload.shift();
                if (!file) {
                    this.channels.push(channel);
                    if(callback instanceof Function){
                        callback();
                    }
                    return;
                }
                doUpload(file,channel, (file,channel,err)=>{
                    this.channels.push(channel);
                    this.startUpload(callback);
                    err&&console.error(err);
                });
            }
        }
    }
    exports.FileUp=FileUp;
}))