from teafiles import *

tf = TeaFile.create("acme.tea", "Time Price Volume Prob Prib", "Data", "QBQBQ", "prices of acme at NYSE", {"decimals": 2, "url": "www.acme.com" })
tf.write(DateTime(2011, 3, 4, 9, 0), 253, 8, 252, 2)
tf.write(DateTime(2011, 3, 4, 9, 0), 253, 8, 252, 2)
tf.close()
