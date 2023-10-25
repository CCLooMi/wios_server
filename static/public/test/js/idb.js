/**
 * Created by chenxianjun on 2022/9/28 14:55:14.
 */
(function (global, factory) {
    'use strict';
    if (typeof exports === 'object' && typeof module !== 'undefined') {
        module.exports = factory();
    } else if (typeof define === 'function' && define.amd) {
        define(factory);
    } else {
        global = global || self;
        global.IDB = factory();
    }
}(this, (function () {
    'use strict';
    var idb=indexedDB;
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
    function final(f) {
        var p = new Promise(function (resolve, reject) {
            var to=setTimeout(()=>{
                clearTimeout(to);
                try{
                    if(!p._args){
                        resolve(f());
                    }else{
                        p._args.then(function (args) {
                            resolve(f.apply(null,args));
                        }).catch(reject);
                    }
                }catch (e) {
                    reject(e);
                }
            },0);
        });
        //支持异步参数
        p.args=function (...args) {
            if(Array.isArray(args[0])){
                args=args[0];
            }
            this._args=Promise.all(args);
            return this;
        };
        return p;
    }
    function reopenWhenVersionChange(db,...reopens) {
        //versionchange只在indexedDB.deleteDatabase(name)时会触发
        var dp=attacheEvent(db).on('versionchange',function (e) {
            dp();
            db.close();
            if(reopens.length){
                if(Array.isArray(reopens[0])){
                    reopens=reopens[0];
                }
                function f(db) {
                    for(var i=0;i<reopens.length;i++){
                        try{reopens[i](db)}catch (e) {};
                    }
                }
                //e.newVersion只在更新时有值，删除时无值
                if(e.newVersion) {
                    f(openIDB(db.name,e.newVersion).onReopen(reopens));
                }else{
                    f(openIDB(db.name).onReopen(reopens));
                }
            }
        }).getDispose();
    }
    function openIDB(name,version) {
        var p = new Promise(function (resolve,reject) {
            var req=idb.open(name,version);
            var dsp=attacheEvent(req)
                .on('upgradeneeded',function (e) {
                    dsp();
                    try{
                        if(p.upgradeneeded){
                            p.upgradeneeded(e.target.transaction);
                        }
                    }catch (e) {
                        reject(e);
                        return;
                    }finally {
                        e.target.result.close();
                    }
                    //when upgrade finished reopenIDB
                    openIDB(name,version)
                        .onReopen(p.reopens)
                        .then(resolve)
                        .catch(reject);
                    p.reopens=[];
                })
                .on('success',function (e) {
                    dsp();
                    reopenWhenVersionChange(e.target.result,p.reopens);
                    p.reopens=[];
                    resolve(e.target.result);
                })
                .on('error',function (e) {
                    dsp();
                    p.reopens=[];
                    reject(e.target.error);
                })
                .on('blocked',function (e) {
                    dsp();
                    p.reopens=[];
                    reject(e);
                })
                .getDispose();
        });
        p.upgrade=function (func) {
            this.upgradeneeded=func;
            return this;
        };
        p.reopens=[];
        p.onReopen=function (...funcs) {
            if(Array.isArray(funcs[0])){
                funcs=funcs[0];
            }
            this.reopens.push.apply(this.reopens,funcs);
            return this;
        };
        return p;
    }
    function getObjectStore(db,name,type) {
        if('rw'==type||'readwrite'==type){
            return db.transaction([name],'readwrite').objectStore(name);
        }
        return db.transaction([name]).objectStore(name);
    }
    function reqToPromise(req) {
        return new Promise(function (resolve, reject) {
            var dsp=attacheEvent(req)
                .on('success',function (e) {
                    dsp();
                    resolve(e);
                })
                .on('error',function (e) {
                    dsp();
                    reject(e.target.error);
                })
                .getDispose();
        })
    }
    function getFromIDB(db) {
        var p = new Promise(function (resolve, reject) {
            final(function (db) {
                var st=getObjectStore(db,p.table);
                var ps=[];
                for(var i=0;i<p.keys.length;i++){
                    ps.push(new Promise(function (resolve, reject) {
                        var req;
                        if(st.indexNames.contains(p.indexName)){
                            req=st.index(p.indexName).getAll(p.keys[i]);
                        }else{
                            req=st.get(p.keys[i]);
                        }
                        var dsp=attacheEvent(req)
                            .on('success',function (e) {
                                dsp();
                                resolve(e.target.result);
                            })
                            .on('error',function (e) {
                                dsp();
                                reject(e.target.error);
                            })
                            .getDispose();
                    }))
                }
                Promise.all(ps).then(resolve).catch(reject);
            }).args(db).catch(reject)
        });
        p.from=function (table) {
            this.table=table;
            return this;
        }
        p.key=function (...keys) {
            if(Array.isArray(keys[0])){
                keys=keys[0];
            }
            this.keys=keys;
            return this;
        }
        p.useIndex=function (name) {
            this.indexName=name+'Idx';
            return this;
        }
        return p;
    }
    function parserRangeArray(range) {
        let args=[],rg=[0,0];
        for(let i=0;i<range.length;i++){
            switch (range[i]) {
                case '(':case '[':
                    rg[0]=range[i];
                    break;
                case ')':case ']':
                    rg[1]=range[i];
                    break;
                default:
                    args.push(range[i]);
                    break;
            }
        }
        if(args.length){
            if(rg[0]&&rg[1]){
                if(args.length>=2){
                    return IDBKeyRange.bound(args[0], args[1], rg[0]=='(', rg[1]==')');
                }
            }else if(rg[0]){
                //x>=?
                return IDBKeyRange.lowerBound(args[0], rg[0]=='(');
            }else if(rg[1]){
                //x<=?
                return IDBKeyRange.upperBound(args[0], rg[1]==')');
            }else if(args.length>1){
                return IDBKeyRange.bound(args[0], args[1]);
            }
            return IDBKeyRange.only(args[0]);
        }
    }
    function rangeFromIDB(db) {
        var p = new Promise(function (resolve, reject) {
            final(function (db) {
                var st=getObjectStore(db,p.table,'rw');
                if(p.indexName&&st.indexNames.contains(p.indexName)){
                    st=st.index(p.indexName);
                }
                var req;
                if(p._range&&p._range.length){
                    var range=parserRangeArray(p._range);
                    req=st.openCursor(range,p.dir||'next');
                }else{
                    req=st.openCursor(p._key,p.dir||'next');
                }
                var result=[];
                var dsp=attacheEvent(req)
                    .on('success',function (e) {
                        var cursor=e.target.result;
                        if(cursor){
                            try{
                                if(p.eachFunc instanceof Function){
                                    var r=p.eachFunc(cursor);
                                    if(r!==undefined){
                                        result.push(r);
                                    }
                                }
                            }catch (ee) {
                                reject(ee);
                            }
                        }else{
                            resolve(result);
                            dsp();
                        }
                    })
                    .on('error',function (e) {
                        dsp();
                        reject(e.target.error);
                    })
                    .getDispose();
            }).args(db).catch(reject);
        });
        p.from=function (table) {
            this.table=table;
            return this;
        };
        p.useIndex=function (idxName) {
            this.indexName=idxName+'Idx';
            return this;
        };
        p.range=function (...range) {
            if(Array.isArray(range[0])){
                range=range[0];
            }
            this._range=range;
            return this;
        };
        p.key=function (key) {
            this._key=key;
            return this;
        };
        p.esc=function () {
            this.dir='next';
            return this;
        };
        p.escTopOne=function () {
            this.dir='nextunique';
            return this;
        };
        p.desc=function () {
            this.dir='prev';
            return this;
        };
        p.descTopOne=function () {
            this.dir='prevunique';
            return this;
        };
        p.each=function (func) {
            this.eachFunc=func;
            return this;
        };
        return p;
    }
    function putToIDB(db) {
        var p = new Promise(function (resolve, reject) {
            final(function (db) {
                var ps=[];
                var st=getObjectStore(db,p.table,'rw');
                for(var i=0;i<p.datas.length;i++){
                    ps.push(reqToPromise(st.put(p.datas[i])));
                }
                Promise.all(ps).then(resolve).catch(reject);
            }).args(db).catch(reject);
        });
        p.to=function (table) {
            this.table=table;
            return this;
        };
        p.data=function (...datas) {
            if(Array.isArray(datas[0])){
                datas=datas[0];
            }
            this.datas=datas;
            return this;
        };
        return p;
    }
    function deleteFromIDB(db) {
        var p = new Promise(function (resolve, reject) {
            final(function (db) {
                var ps=[];
                var st=getObjectStore(db,p.table,'rw');
                for(var i=0;i<p.keys.length;i++){
                    ps.push(reqToPromise(st.delete(p.keys[i])));
                }
                Promise.all(ps).then(resolve).catch(reject);
            }).args(db).catch(reject);
        });
        p.from=function (table) {
            this.table=table;
            return this;
        };
        p.key=function (...keys) {
            if(Array.isArray(keys[0])){
                keys=keys[0];
            }
            this.keys=keys;
            return this;
        };
        return p;
    }
    function clearFromIDB(db) {
        var p = new Promise(function (resolve, reject) {
            final(function (db) {
                var st=getObjectStore(db,p.table,'rw');
                var dsp=attacheEvent(st.clear())
                    .on('success',function (e) {
                        dsp();
                        resolve(e);
                    })
                    .on('error',function (e) {
                        dsp();
                        reject(e.target.error);
                    })
                    .getDispose();
            }).args(db).catch(reject);
        });
        p.from=function (table) {
            this.table=table;
            return this;
        };
        return p;
    }
    function countFromIDB(db) {
        var p = new Promise(function (resolve, reject) {
            final(function (db) {
                var st=getObjectStore(db,p.table);
                var req;
                if(st.indexNames.contains(p.indexName)){
                    req=st.index(p.indexName).count(p._key);
                }else{
                    req=st.count(p._key);
                }
                var dsp= attacheEvent(req)
                    .on('success',function (e) {
                        dsp();
                        resolve(e.target.result);
                    })
                    .on('error',function (e) {
                        dsp();
                        reject(e.target.error);
                    })
                    .getDispose();
            }).args(db).catch(reject);
        });
        p.from=function (table) {
            this.table=table;
            return this;
        };
        p.key=function (key) {
            this._key=key;
            return this;
        };
        p.useIndex=function (name) {
            this.indexName=name+'Idx';
            return this;
        };
        return p;
    }
    function removeIDB(name) {
        return reqToPromise(idb.deleteDatabase(name));
    }
    function normalizeObj(obj) {
        var keys=Object.keys(obj).sort();
        var newObj={};
        for(var i=0;i<keys.length;i++){
            newObj[keys[i]]=obj[keys[i]];
        }
        return newObj;
    }
    function openOrCreateIDB(name,schema) {
        schema=normalizeObj(schema);
        var options={};
        for(var p in schema){
            var opt={};
            var ss=schema[p].split(',');
            for(var i=0;i<ss.length;i++){
                var sci=ss[i].split(' ');
                switch (sci[1]) {
                    case '+':case 'k':case 'K':
                        if(sci[1]==='+'||sci[2]==='+'){
                            opt.autoIncrement=true;
                        }else if(sci[1]==='k'||sci[1]==='K'||sci[2]==='k'||sci[2]==='K'){
                            opt.keyPath=sci[0];
                        }
                        ss.splice(i--,1);
                        break;
                    default:
                        ss[i]=sci;
                        break;
                }
            }
            schema[p]=ss;
            options[p]=opt;
        }

        var pm=new Promise(function (resolve, reject) {
            openIDB(name).then(function (db) {
                var upgradeFunc=[];
                var sts=db.objectStoreNames;
                for(var i=0;i<sts.length;i++){
                    //remove objStore not in schema
                    if(!schema[sts[i]]){
                        upgradeFunc.push((function (name) {
                            return function (ts) {
                                return ts.db.deleteObjectStore(name);
                            };
                        })(sts[i]));
                        continue;
                    }
                    //check update
                    var st=getObjectStore(db,sts[i]);
                    var opt=options[st.name];
                    //objStore changed,we need delete it first and then recreate it;
                    if((opt.autoIncrement||false)!=st.autoIncrement||opt.keyPath!=st.keyPath){
                        //TODO old objStore data will lost! maybe we can backup first!
                        upgradeFunc.push((function (table,opt) {
                            return function (ts) {
                                ts.db.deleteObjectStore(table);
                                ts.db.createObjectStore(table,opt);
                            };
                        })(st.name,opt));
                    }
                    //check index
                    var sc=schema[sts[i]];
                    var idxNames={};//用于判断st.indexNames中是否有需要删除的index
                    for(var j=0;j<sc.length;j++){
                        var sci=sc[j];
                        var idxName=sci[0]+'Idx';
                        idxNames[idxName]=true;
                        var idx=st.indexNames.contains(idxName)?st.index(idxName):null;
                        switch (sci[1]) {
                            case 'u':case 'U':
                                if(idx&&idx.unique){
                                    break;
                                }
                                upgradeFunc.push((function (table,keyPath,deleteIdx) {
                                    return function (ts) {
                                        var st=ts.objectStore(table);
                                        if(deleteIdx){
                                            st.deleteIndex(keyPath+'Idx');
                                        }
                                        st.createIndex(keyPath+'Idx',keyPath,{unique:true});
                                    };
                                })(st.name,sci[0],idx&&!idx.unique));
                                break;
                            case '[]':
                                if(idx&&idx.multiEntry){
                                    break;
                                }
                                upgradeFunc.push((function (table,keyPath,deleteIdx) {
                                    return function (ts) {
                                        var st=ts.objectStore(table);
                                        if(deleteIdx){
                                            st.deleteIndex(keyPath+'Idx');
                                        }
                                        st.createIndex(keyPath+'Idx',keyPath,{multiEntry:true});
                                    };
                                })(st.name,sci[0],idx&&!idx.multiEntry));
                                break;
                            default:
                                if(idx&&!idx.unique&&!idx.multiEntry){
                                    break;
                                }
                                upgradeFunc.push((function (table,keyPath,deleteIdx) {
                                    return function (ts) {
                                        var st=ts.objectStore(table);
                                        if(deleteIdx){
                                            st.deleteIndex(keyPath+'Idx');
                                        }
                                        st.createIndex(keyPath+'Idx',keyPath);
                                    };
                                })(st.name,sci[0],idx&&(idx.unique||idx.multiEntry)));
                                break;
                        }
                    }
                    var removeList=[];
                    for(var j=0;j<st.indexNames.length;j++){
                        if(!idxNames[st.indexNames[j]]){
                            removeList.push(st.indexNames[j]);
                        }
                    }
                    if(removeList.length){
                        upgradeFunc.push((function (table,rList) {
                            return function (ts) {
                                var st=ts.objectStore(table);
                                for(var j=0;j<rList.length;j++){
                                    st.deleteIndex(rList[j]);
                                }
                            };
                        })(st.name,removeList));
                    }
                    delete schema[st.name];
                }
                //create new objStore and index
                for(var p in schema){
                    var sc=schema[p];
                    var opt=options[p];
                    upgradeFunc.push((function (table,options) {
                        return function (ts) {
                            ts.db.createObjectStore(table,options);
                        };
                    })(p,opt));
                    //create new index in objStore
                    for(var i=0;i<sc.length;i++){
                        var sci=sc[i];
                        switch (sci[1]) {
                            case 'u':case 'U':
                                upgradeFunc.push((function (table,keyPath) {
                                    return function (ts) {
                                        var st=ts.objectStore(table);
                                        st.createIndex(keyPath+'Idx',keyPath,{unique:true});
                                    };
                                })(p,sci[0]));
                                break;
                            case '[]':
                                upgradeFunc.push((function (table,keyPath) {
                                    return function (ts) {
                                        var st=ts.objectStore(table);
                                        st.createIndex(keyPath+'Idx',keyPath,{multiEntry:true});
                                    };
                                })(p,sci[0]));
                                break;
                            default:
                                upgradeFunc.push((function (table,keyPath) {
                                    return function (ts) {
                                        var st=ts.objectStore(table);
                                        st.createIndex(keyPath+'Idx',keyPath);
                                    };
                                })(p,sci[0]));
                                break;
                        }
                    }
                    delete schema[p];
                }
                if(upgradeFunc.length){
                    openIDB(name,Date.now())
                        .upgrade(function (ts) {
                            var ps=[];
                            for(var i=0;i<upgradeFunc.length;i++){
                                var req=upgradeFunc[i](ts);
                                if(!req){
                                    continue;
                                }
                                ps.push(reqToPromise(req));
                            }
                            return Promise.all(ps);
                        })
                        .onReopen(pm.reopens)
                        .then(resolve)
                        .catch(reject);
                }else{
                    reopenWhenVersionChange(db,pm.reopens);
                    resolve(db);
                }
                pm.reopens=[];
            }).catch(reject);
        });
        pm.reopens=[];
        pm.onReopen=function (...funcs) {
            if(Array.isArray(funcs[0])){
                funcs=funcs[0];
            }
            this.reopens.push.apply(this.reopens,funcs);
            return this;
        };
        return pm;
    }

    var idbSet={};
    function initDB(idb,name,schema) {
        return new Promise(function (resolve, reject) {
            var setDB=function(db){
                idbSet[name]=db;
                setThisDB(db);
            };
            var setThisDB=function(db){
                idb.db=db;
                resolve(db);
            };
            if(schema){
                if(!idbSet[name]){
                    idbSet[name]=openOrCreateIDB(name,schema).onReopen(setDB);
                    idbSet[name].then(setDB);
                }else if(idbSet[name] instanceof Promise){
                    idbSet[name].onReopen(setThisDB).then(setThisDB);
                }
            }else{
                if(!idbSet[name]){
                    idbSet[name]=openIDB(name).onReopen(setDB);
                    idbSet[name].then(setDB);
                }else if(idbSet[name] instanceof Promise){
                    idbSet[name].onReopen(setThisDB).then(setThisDB);
                }
            }
        });
    }
    function IDB(name,schema) {
        this.name=name;
        initDB(this,name,schema);
        this.db=idbSet[name];
    }
    IDB.prototype={
        reopen:function () {
            delete idbSet[this.name];
            delete this.db;
            return initDB(this,this.name);
        },
        get:function (table,...keys) {
            if(Array.isArray(keys[0])){
                keys=keys[0];
            }
            return getFromIDB(this.db).from(table).key(keys);
        },
        range:function (table,...range) {
            if(Array.isArray(range[0])){
                range=range[0];
            }
            return rangeFromIDB(this.db).from(table).range(range);
        },
        foreach:function (table,key) {
            return rangeFromIDB(this.db).from(table).key(key);
        },
        put:function (table,...datas) {
            if(Array.isArray(datas[0])){
                datas=datas[0];
            }
            return putToIDB(this.db).to(table).data(datas);
        },
        count:function (table,key) {
            return countFromIDB(this.db).from(table).key(key);
        },
        delete:function (table,...keys) {
            if(Array.isArray(keys[0])){
                keys=keys[0];
            }
            return deleteFromIDB(this.db).from(table).key(keys);
        },
        byPage:function (table,pgObj) {
            return new Promise( (resolve, reject)=>{
                var kvObj={};
                for(var k in pgObj) {
                    if(k=='page' || k== 'pageSize'){
                        continue;
                    }
                    kvObj[k]=new RegExp('.*'+pgObj[k]+'.*');
                }
                var page = pgObj.page;
                var pageSize = pgObj.pageSize;
                var start = (page - 1) * pageSize;
                var end = start + pageSize;
                var i = 0;
                this.foreach(table).each(function (c) {
                    var v=c.value;
                    c.continue();
                    for(var k in kvObj){
                        if(v.hasOwnProperty(k)){
                            if(!kvObj[k].test(v[k])){
                                return;
                            }
                        }
                    }
                    try{
                        if (i >= start && i < end) {
                            return v;
                        }
                    }finally {
                        i++;
                    }
                }).then(values => {
                    resolve({ total: i, data: values });
                },reject);
            });
        },
        tables:function () {
            return final(function (db) {
                return db.objectStoreNames;
            }).args(this.db);
        },
        clear:function (table) {
            return clearFromIDB(this.db).from(table);
        }
    };
    IDB.reqToPromise=reqToPromise;
    return IDB;
})));