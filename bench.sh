for i in {1..1}
do
	curl -X POST -d "{\"ts\":1609310997, \"name\":\"jo\"}" http://localhost:8080/user/0x970E8128AB834E8EAC17Ab8E3812F010678CF791?sig=0x15ce17f60e6825a4d5556867c30d3bc823f9f2dd0d55aa845a816f4518a081ca5e2c9fea9ec552e861d015306c6c7c4132135e97b0e695e01c751c51e5e7075d01
	echo ""
	curl -X POST -d "{\"ts\":1609310997, \"name\":\"jo\", \"age\":30}" http://localhost:8080/user/0x970E8128AB834E8EAC17Ab8E3812F010678CF791?sig=0x2e9d7610f3611be41ded12d595285530387522fbc60ca691811adb4306996bc6275e80e292041dd4eecc65bc1d6e53e410eb7390e0fca6fd7d2c378a6f3efa7000
	echo ""
	curl -X POST -d "{\"ts\":1609310997, \"addr\":\"0x2a2a0667f9cbf4055e48eaf0d5b40304b8822184\"}" http://localhost:8080/favor/0x970E8128AB834E8EAC17Ab8E3812F010678CF791?sig=0xab133a9294a829e9023264cd146a8ab56d9ff8ead591cc7d218d5cbd1513f8a73fe7c6666b86f8aaffef8ba6f3cef4bfaebe4a7502df052803965440da7baa7300
	echo ""
	curl -X POST -d "{\"ts\":1609310997, \"addr\":\"0x970E8128AB834E8EAC17Ab8E3812F010678CF791\"}" http://localhost:8080/favor/0x970E8128AB834E8EAC17Ab8E3812F010678CF791?sig=0x665587580842105a1379cc214ee722a2b0e49d5a08ec92cbf974aa9f49b8716361937de3385eaeda010e65cf06c5cb2581a82210ec34db34573fd9d4c0527dc600
	echo ""
	curl -X POST -d "{\"ts\":1609310997, \"addr\":\"0x564286362092d8e7936f0549571a803b203aaced\"}" http://localhost:8080/favor/0x970E8128AB834E8EAC17Ab8E3812F010678CF791?sig=0x3994ff63e3886d7d811fbcd62c14303bbc62cea585efa525e90835b79f33ace13b6fd706581b0cbec0c6c5f42baab286b9af5ba8ca350687297bad2bbb5e052300
	echo ""
	#curl -X DELETE -d "{\"ts\":1609310997, \"addr\":\"0x564286362092d8e7936f0549571a803b203aaced\"}" http://localhost:8080/favor/0x970E8128AB834E8EAC17Ab8E3812F010678CF791?sig=0x3994ff63e3886d7d811fbcd62c14303bbc62cea585efa525e90835b79f33ace13b6fd706581b0cbec0c6c5f42baab286b9af5ba8ca350687297bad2bbb5e052300
	echo ""
done
