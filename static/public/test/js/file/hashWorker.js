/**
 * Created by chenxianjun on 2023/4/27 16:55:31.
 */
let csm = [];
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
function bytesToWords(a) {
    for (var b = [], c = 0, d = 0; c < a.length; c++, d += 8) {
        b[d >>> 5] |= a[c] << 24 - d % 32;
    }
    return b
}
function wordsToBytes(a) {
    for (var b = [], c = 0; c < a.length * 32; c += 8) {
        b.push(a[c >>> 5] >>> 24 - c % 32 & 255);
    }
    return b
}
let algo = null;//CryptoJS.algo.SHA256.create();
let firstHash = null;
let step = 0;
let hashDB = new IDB('hash', {
    'hash': 'id k'
})
self.addEventListener('message', function (event) {
    var d = event.data;
    try {
        //init algo
        if (d.algo) {
            firstHash = null;
            step = 0;
            algo = CryptoJS.algo[d.algo].create();
        }
        let block = d.block;
        let output = {'block': block};
        if (!block.stop) {
            var ps = [];
            let message = d.msg;
            if (block.end === block.fileSize) {
                algo.update(CryptoJS.lib.WordArray.create(message));
                self.hash = algo.finalize();
                // output.result = self.hash.toString(CryptoJS.enc.Hex);
                output.result = bytesToHex(wordsToBytes(self.hash.words));
                if (firstHash) {
                    ps.push(hashDB.put('hash', {id: firstHash, sha: output.result}));
                }
            } else {
                algo.update(CryptoJS.lib.WordArray.create(message));
                step++;
                if (!firstHash && step == 3) {
                    let copyHash = algo._hash.words.concat();
                    firstHash = bytesToHex(wordsToBytes(copyHash));
                    ps.push(hashDB.get('hash', firstHash));
                }
            }
            if (ps.length) {
                Promise.all(ps).then(r => {
                    if (r.length == 1) {
                        r = r[0][0];
                    } else {
                        r = r[1][0];
                    }
                    if (r && r.sha) {
                        block.end = block.fileSize;
                        output.result = r.sha;
                    }
                    self.postMessage(output);
                });
            } else {
                self.postMessage(output);
            }
        }
    } catch (e) {
        self.postMessage({error: e});
    }
}, false);