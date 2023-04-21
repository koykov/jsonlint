package jsonlint

import "testing"

var (
	jsvGood = [][]byte{
		[]byte(`{"glossary":{"title":"example glossary","GlossDiv":{"title":"S","GlossList":{"GlossEntry":{"ID":"SGML","SortAs":"SGML","GlossTerm":"Standard Generalized Markup Language","Acronym":"SGML","Abbrev":"ISO 8879:1986","GlossDef":{"para":"A meta-markup language, used to create markup languages such as DocBook.","GlossSeeAlso":["GML","XML"]},"GlossSee":"markup"}}}}}`),
		[]byte(`{"menu":{"id":"file","value":"File","popup":{"menuitem":[{"value":"New","onclick":"CreateNewDoc()"},{"value":"Open","onclick":"OpenDoc()"},{"value":"Close","onclick":"CloseDoc()"}]}}}`),
		[]byte(`{"widget":{"debug":"on","window":{"title":"Sample Konfabulator Widget","name":"main_window","width":500,"height":500},"image":{"src":"Images/Sun.png","name":"sun1","hOffset":250,"vOffset":250,"alignment":"center"},"text":{"data":"Click Here","size":36,"style":"bold","name":"text1","hOffset":250,"vOffset":100,"alignment":"center","onMouseUp":"sun1.opacity = (sun1.opacity / 100) * 90;"}}}`),
		[]byte(`{"web-app":{"servlet":[{"servlet-name":"cofaxCDS","servlet-class":"org.cofax.cds.CDSServlet","init-param":{"configGlossary:installationAt":"Philadelphia, PA","configGlossary:adminEmail":"ksm@pobox.com","configGlossary:poweredBy":"Cofax","configGlossary:poweredByIcon":"/images/cofax.gif","configGlossary:staticPath":"/content/static","templateProcessorClass":"org.cofax.WysiwygTemplate","templateLoaderClass":"org.cofax.FilesTemplateLoader","templatePath":"templates","templateOverridePath":"","defaultListTemplate":"listTemplate.htm","defaultFileTemplate":"articleTemplate.htm","useJSP":false,"jspListTemplate":"listTemplate.jsp","jspFileTemplate":"articleTemplate.jsp","cachePackageTagsTrack":200,"cachePackageTagsStore":200,"cachePackageTagsRefresh":60,"cacheTemplatesTrack":100,"cacheTemplatesStore":50,"cacheTemplatesRefresh":15,"cachePagesTrack":200,"cachePagesStore":100,"cachePagesRefresh":10,"cachePagesDirtyRead":10,"searchEngineListTemplate":"forSearchEnginesList.htm","searchEngineFileTemplate":"forSearchEngines.htm","searchEngineRobotsDb":"WEB-INF/robots.db","useDataStore":true,"dataStoreClass":"org.cofax.SqlDataStore","redirectionClass":"org.cofax.SqlRedirection","dataStoreName":"cofax","dataStoreDriver":"com.microsoft.jdbc.sqlserver.SQLServerDriver","dataStoreUrl":"jdbc:microsoft:sqlserver://LOCALHOST:1433;DatabaseName=goon","dataStoreUser":"sa","dataStorePassword":"dataStoreTestQuery","dataStoreTestQuery":"SET NOCOUNT ON;select test='test';","dataStoreLogFile":"/usr/local/tomcat/logs/datastore.log","dataStoreInitConns":10,"dataStoreMaxConns":100,"dataStoreConnUsageLimit":100,"dataStoreLogLevel":"debug","maxUrlLength":500}},{"servlet-name":"cofaxEmail","servlet-class":"org.cofax.cds.EmailServlet","init-param":{"mailHost":"mail1","mailHostOverride":"mail2"}},{"servlet-name":"cofaxAdmin","servlet-class":"org.cofax.cds.AdminServlet"},{"servlet-name":"fileServlet","servlet-class":"org.cofax.cds.FileServlet"},{"servlet-name":"cofaxTools","servlet-class":"org.cofax.cms.CofaxToolsServlet","init-param":{"templatePath":"toolstemplates/","log":1,"logLocation":"/usr/local/tomcat/logs/CofaxTools.log","logMaxSize":"","dataLog":1,"dataLogLocation":"/usr/local/tomcat/logs/dataLog.log","dataLogMaxSize":"","removePageCache":"/content/admin/remove?cache=pages&id=","removeTemplateCache":"/content/admin/remove?cache=templates&id=","fileTransferFolder":"/usr/local/tomcat/webapps/content/fileTransferFolder","lookInContext":1,"adminGroupID":4,"betaServer":true}}],"servlet-mapping":{"cofaxCDS":"/","cofaxEmail":"/cofaxutil/aemail/*","cofaxAdmin":"/admin/*","fileServlet":"/static/*","cofaxTools":"/tools/*"},"taglib":{"taglib-uri":"cofax.tld","taglib-location":"/WEB-INF/tlds/cofax.tld"}}}`),
		[]byte(`{"menu":{"header":"SVG Viewer","items":[{"id":"Open"},{"id":"OpenNew","label":"Open New"},null,{"id":"ZoomIn","label":"Zoom In"},{"id":"ZoomOut","label":"Zoom Out"},{"id":"OriginalView","label":"Original View"},null,{"id":"Quality"},{"id":"Pause"},{"id":"Mute"},null,{"id":"Find","label":"Find..."},{"id":"FindAgain","label":"Find Again"},{"id":"Copy"},{"id":"CopyAgain","label":"Copy Again"},{"id":"CopySVG","label":"Copy SVG"},{"id":"ViewSVG","label":"View SVG"},{"id":"ViewSource","label":"View Source"},{"id":"SaveAs","label":"Save As"},null,{"id":"Help"},{"id":"About","label":"About Adobe CVG Viewer..."}]}}`),
	}
	jsvBad = [][]byte{
		// empty source
		[]byte(``),
		// unparsed tail
		[]byte(`{"menu":{"id":"file","value":"File","popup":{"menuitem":[{"value":"New","onclick":"CreateNewDoc()"},{"value":"Open","onclick":"OpenDoc()"},{"value":"Close","onclick":"CloseDoc()"}]}}},"foo"`),
		// unexpected identifier
		[]byte(`{"menu":{"id":"file","value","popup":{"menuitem":[{"value":"New","onclick":"CreateNewDoc()"},{"value":"Open","onclick":"OpenDoc()"},{"value":"Close","onclick":"CloseDoc()"}]}}}`),
		// unexpected EOF
		[]byte(`{"menu":{"id":"file","value":"File","popup":{"menuitem":[{"value":"New","onclick":"CreateNewDoc()"},{"value":"Open","onclick":"OpenDoc()"},{"value":"Close","onclick":"CloseDoc()"}]}`),
		// unclosed string
		[]byte(`{"menu":{"id":"file","value":"File","popup":{"menuitem":[{"value":"New","onclick":"CreateNewDoc()"},{"value":"Open","onclick":"OpenDoc},{"value":"Close","onclick":"CloseDoc()"}]}}}`),
		// empty array item
		[]byte(`{"menu":{"id":"file","value":"File","popup":{"menuitem":[{"value":"New","onclick":"CreateNewDoc()"},{"value":"Open","onclick":"OpenDoc()"},,{"value":"Close","onclick":"CloseDoc()"}]}}}`),
	}
)

func TestJsonlint(t *testing.T) {
	t.Run("ValidateGood", func(t *testing.T) {
		var (
			o   int
			err error
		)
		for _, jsv := range jsvGood {
			o, err = Validate(jsv)
			if err != nil {
				t.Error(err)
			}
			if o < len(jsv) {
				t.Error("unparsed tail", len(jsv)-o, "bytes")
			}
		}
	})
	t.Run("empty", func(t *testing.T) {
		_, err := Validate(jsvBad[0])
		if err != ErrEmptySrc {
			t.Error("need", ErrEmptySrc, "got", err)
		}
	})
	t.Run("unparsed tail", func(t *testing.T) {
		o, err := Validate(jsvBad[1])
		if err != ErrUnparsedTail {
			t.Error("need", ErrUnparsedTail, "got", err)
		}
		if o != 183 {
			t.Error("bad offset", o, "need", 183)
		}
	})
	t.Run("unexpected id", func(t *testing.T) {
		o, err := Validate(jsvBad[2])
		if err != ErrUnexpId {
			t.Error("need", ErrUnexpId, "got", err)
		}
		if o != 28 {
			t.Error("bad offset", o, "need", 28)
		}
	})
	t.Run("unexpected EOF", func(t *testing.T) {
		o, err := Validate(jsvBad[3])
		if err != ErrUnexpEOF {
			t.Error("need", ErrUnexpEOF, "got", err)
		}
		if o != 181 {
			t.Error("bad offset", o, "need", 181)
		}
	})
	t.Run("unclosed string", func(t *testing.T) {
		o, err := Validate(jsvBad[4])
		if err != ErrUnexpId {
			t.Error("need", ErrUnexpId, "got", err)
		}
		if o != 138 {
			t.Error("bad offset", o, "need", 138)
		}
	})
	t.Run("empty array item", func(t *testing.T) {
		o, err := Validate(jsvBad[5])
		if err != ErrUnexpId {
			t.Error("need", ErrUnexpId, "got", err)
		}
		if o != 139 {
			t.Error("bad offset", o, "need", 139)
		}
	})
}

func BenchmarkJsonlint(b *testing.B) {
	b.Run("ok", func(b *testing.B) {
		var (
			o   int
			err error
		)
		b.ResetTimer()
		b.ReportAllocs()
		l := len(jsvGood[0])
		for i := 0; i < b.N; i++ {
			o, err = Validate(jsvGood[0])
			if err != nil {
				b.Error(err)
			}
			if o < l {
				b.Error("unparsed tail", l-o, "bytes")
			}
		}
	})
	b.Run("unparsed tail", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			o, err := Validate(jsvBad[1])
			if err != ErrUnparsedTail {
				b.Error("need", ErrUnparsedTail, "got", err)
			}
			if o != 183 {
				b.Error("bad offset", o, "need", 183)
			}
		}
	})
	b.Run("unexpected id", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			o, err := Validate(jsvBad[2])
			if err != ErrUnexpId {
				b.Error("need", ErrUnexpId, "got", err)
			}
			if o != 28 {
				b.Error("bad offset", o, "need", 28)
			}
		}
	})
	b.Run("unexpected EOF", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			o, err := Validate(jsvBad[3])
			if err != ErrUnexpEOF {
				b.Error("need", ErrUnexpEOF, "got", err)
			}
			if o != 181 {
				b.Error("bad offset", o, "need", 181)
			}
		}
	})
	b.Run("unclosed string", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			o, err := Validate(jsvBad[4])
			if err != ErrUnexpId {
				b.Error("need", ErrUnexpId, "got", err)
			}
			if o != 138 {
				b.Error("bad offset", o, "need", 138)
			}
		}
	})
	b.Run("empty array item", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			o, err := Validate(jsvBad[5])
			if err != ErrUnexpId {
				b.Error("need", ErrUnexpId, "got", err)
			}
			if o != 139 {
				b.Error("bad offset", o, "need", 139)
			}
		}
	})
}
