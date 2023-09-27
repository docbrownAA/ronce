deps::
	npm install

build::
	rm -f ./src/notification-api/assets/subject;\
	for file in src/mjml/*.mjml ; do \
		name=`basename $$file .mjml` ;\
		echo `sed -n 1p $$file` >> ./src/notification-api/assets/subject;\
		mjml $$file -o ./src/notification-api/assets/$$name.html > /dev/null;\
		echo building $$name.html ;\
	done

test:: deps