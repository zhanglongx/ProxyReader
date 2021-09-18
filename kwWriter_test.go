package main

import (
	"bytes"
	"log"
	"os"
	"testing"
)

func Test_kwWriter_Close(t *testing.T) {
	type fields struct {
		Path string
		buf  bytes.Buffer
	}
	tests := []struct {
		name    string
		fields  fields
		vector  string
		wantErr bool
	}{
		{name: "Good", fields: fields{Path: "tmp"}, vector: "{\"status\":true,\"data\":{\"need_open_comment\":1,\"content\":\"<section></section>\",\"content_source_url\":\"https://m.joudou.com/invwinner#/winner?page=term&termid=1\",\"digest\":\"挥别四年一轮的二二九\",\"title\":\"【刊外四十】玩转双增长\",\"author\":\"骑行客\",\"only_fans_can_comment\":0,\"url\":\"http://mp.weixin.qq.com/s?__biz=Mzg3MTIzNTkyMw==&mid=100000703&idx=1&sn=5d92abf3f1c31eb52303c301d15bc016&chksm=4e80eb8379f76295c63cc75b48f890991c7238c6dcdf749dc060a09f70fbe8dbb4a9de0d894d#rd\",\"ext_info\":{\"read_count\":0,\"like_count\":0,\"name\":\"霄云笔谈\",\"prdname\":\"刊外\",\"jumpurl\":\"https://m.joudou.com/invwinner-kanwai/6422658b16dd4e0fb3cb2ad4ab7b648a\",\"mp_name\":\"骑行夜幕统计客\",\"mp_url\":\"https://mp.weixin.qq.com/mp/profile_ext?action=home&__biz=Mzg3MTIzNTkyMw==&scene=110#wechat_redirect\",\"icon\":\"https://static.joudou.com/wechat/dknight_avatar.png\"},\"thumb_media_id\":\"PNE0wI8i_2_xJurtZ7ztbj-wBkXRjT6H72nX-XM9gnw\",\"thumb_url\":\"http://mmbiz.qpic.cn/mmbiz_jpg/BOQXE8706a8mic3ibMiaNy0XfMam07N0JuBvj4jv2E3L2eFg8FibA78ehx4iayAXetECiaU0preQRPjevZoKXDaJTibOQ/0?wx_fmt=jpeg\",\"releasedate\":\"2020-02-29T16:35:33Z\",\"show_cover_pic\":0}}",
			wantErr: false},
		{name: "No Data", fields: fields{Path: "tmp"}, vector: "{\"status\":true,\"dat\":{\"need_open_comment\":1,\"content\":\"<section></section>\",\"content_source_url\":\"https://m.joudou.com/invwinner#/winner?page=term&termid=1\",\"digest\":\"挥别四年一轮的二二九\",\"title\":\"【刊外四十】玩转双增长\",\"author\":\"骑行客\",\"only_fans_can_comment\":0,\"url\":\"http://mp.weixin.qq.com/s?__biz=Mzg3MTIzNTkyMw==&mid=100000703&idx=1&sn=5d92abf3f1c31eb52303c301d15bc016&chksm=4e80eb8379f76295c63cc75b48f890991c7238c6dcdf749dc060a09f70fbe8dbb4a9de0d894d#rd\",\"ext_info\":{\"read_count\":0,\"like_count\":0,\"name\":\"霄云笔谈\",\"prdname\":\"刊外\",\"jumpurl\":\"https://m.joudou.com/invwinner-kanwai/6422658b16dd4e0fb3cb2ad4ab7b648a\",\"mp_name\":\"骑行夜幕统计客\",\"mp_url\":\"https://mp.weixin.qq.com/mp/profile_ext?action=home&__biz=Mzg3MTIzNTkyMw==&scene=110#wechat_redirect\",\"icon\":\"https://static.joudou.com/wechat/dknight_avatar.png\"},\"thumb_media_id\":\"PNE0wI8i_2_xJurtZ7ztbj-wBkXRjT6H72nX-XM9gnw\",\"thumb_url\":\"http://mmbiz.qpic.cn/mmbiz_jpg/BOQXE8706a8mic3ibMiaNy0XfMam07N0JuBvj4jv2E3L2eFg8FibA78ehx4iayAXetECiaU0preQRPjevZoKXDaJTibOQ/0?wx_fmt=jpeg\",\"releasedate\":\"2020-02-29T16:35:33Z\",\"show_cover_pic\":0}}",
			wantErr: true},
		{name: "No Status", fields: fields{Path: "tmp"}, vector: "{\"stat\":true,\"data\":{\"need_open_comment\":1,\"content\":\"<section></section>\",\"content_source_url\":\"https://m.joudou.com/invwinner#/winner?page=term&termid=1\",\"digest\":\"挥别四年一轮的二二九\",\"title\":\"【刊外四十】玩转双增长\",\"author\":\"骑行客\",\"only_fans_can_comment\":0,\"url\":\"http://mp.weixin.qq.com/s?__biz=Mzg3MTIzNTkyMw==&mid=100000703&idx=1&sn=5d92abf3f1c31eb52303c301d15bc016&chksm=4e80eb8379f76295c63cc75b48f890991c7238c6dcdf749dc060a09f70fbe8dbb4a9de0d894d#rd\",\"ext_info\":{\"read_count\":0,\"like_count\":0,\"name\":\"霄云笔谈\",\"prdname\":\"刊外\",\"jumpurl\":\"https://m.joudou.com/invwinner-kanwai/6422658b16dd4e0fb3cb2ad4ab7b648a\",\"mp_name\":\"骑行夜幕统计客\",\"mp_url\":\"https://mp.weixin.qq.com/mp/profile_ext?action=home&__biz=Mzg3MTIzNTkyMw==&scene=110#wechat_redirect\",\"icon\":\"https://static.joudou.com/wechat/dknight_avatar.png\"},\"thumb_media_id\":\"PNE0wI8i_2_xJurtZ7ztbj-wBkXRjT6H72nX-XM9gnw\",\"thumb_url\":\"http://mmbiz.qpic.cn/mmbiz_jpg/BOQXE8706a8mic3ibMiaNy0XfMam07N0JuBvj4jv2E3L2eFg8FibA78ehx4iayAXetECiaU0preQRPjevZoKXDaJTibOQ/0?wx_fmt=jpeg\",\"releasedate\":\"2020-02-29T16:35:33Z\",\"show_cover_pic\":0}}",
			wantErr: false},
		{name: "No Title", fields: fields{Path: "tmp"}, vector: "{\"status\":true,\"data\":{\"need_open_comment\":1,\"content\":\"<section></section>\",\"content_source_url\":\"https://m.joudou.com/invwinner#/winner?page=term&termid=1\",\"digest\":\"挥别四年一轮的二二九\",\"tle\":\"【刊外四十】玩转双增长\",\"author\":\"骑行客\",\"only_fans_can_comment\":0,\"url\":\"http://mp.weixin.qq.com/s?__biz=Mzg3MTIzNTkyMw==&mid=100000703&idx=1&sn=5d92abf3f1c31eb52303c301d15bc016&chksm=4e80eb8379f76295c63cc75b48f890991c7238c6dcdf749dc060a09f70fbe8dbb4a9de0d894d#rd\",\"ext_info\":{\"read_count\":0,\"like_count\":0,\"name\":\"霄云笔谈\",\"prdname\":\"刊外\",\"jumpurl\":\"https://m.joudou.com/invwinner-kanwai/6422658b16dd4e0fb3cb2ad4ab7b648a\",\"mp_name\":\"骑行夜幕统计客\",\"mp_url\":\"https://mp.weixin.qq.com/mp/profile_ext?action=home&__biz=Mzg3MTIzNTkyMw==&scene=110#wechat_redirect\",\"icon\":\"https://static.joudou.com/wechat/dknight_avatar.png\"},\"thumb_media_id\":\"PNE0wI8i_2_xJurtZ7ztbj-wBkXRjT6H72nX-XM9gnw\",\"thumb_url\":\"http://mmbiz.qpic.cn/mmbiz_jpg/BOQXE8706a8mic3ibMiaNy0XfMam07N0JuBvj4jv2E3L2eFg8FibA78ehx4iayAXetECiaU0preQRPjevZoKXDaJTibOQ/0?wx_fmt=jpeg\",\"releasedate\":\"2020-02-29T16:35:33Z\",\"show_cover_pic\":0}}",
			wantErr: true},
		{name: "No content", fields: fields{Path: "tmp"}, vector: "{\"status\":true,\"data\":{\"need_open_comment\":1,\"cont\":\"<section></section>\",\"content_source_url\":\"https://m.joudou.com/invwinner#/winner?page=term&termid=1\",\"digest\":\"挥别四年一轮的二二九\",\"title\":\"【刊外四十】玩转双增长\",\"author\":\"骑行客\",\"only_fans_can_comment\":0,\"url\":\"http://mp.weixin.qq.com/s?__biz=Mzg3MTIzNTkyMw==&mid=100000703&idx=1&sn=5d92abf3f1c31eb52303c301d15bc016&chksm=4e80eb8379f76295c63cc75b48f890991c7238c6dcdf749dc060a09f70fbe8dbb4a9de0d894d#rd\",\"ext_info\":{\"read_count\":0,\"like_count\":0,\"name\":\"霄云笔谈\",\"prdname\":\"刊外\",\"jumpurl\":\"https://m.joudou.com/invwinner-kanwai/6422658b16dd4e0fb3cb2ad4ab7b648a\",\"mp_name\":\"骑行夜幕统计客\",\"mp_url\":\"https://mp.weixin.qq.com/mp/profile_ext?action=home&__biz=Mzg3MTIzNTkyMw==&scene=110#wechat_redirect\",\"icon\":\"https://static.joudou.com/wechat/dknight_avatar.png\"},\"thumb_media_id\":\"PNE0wI8i_2_xJurtZ7ztbj-wBkXRjT6H72nX-XM9gnw\",\"thumb_url\":\"http://mmbiz.qpic.cn/mmbiz_jpg/BOQXE8706a8mic3ibMiaNy0XfMam07N0JuBvj4jv2E3L2eFg8FibA78ehx4iayAXetECiaU0preQRPjevZoKXDaJTibOQ/0?wx_fmt=jpeg\",\"releasedate\":\"2020-02-29T16:35:33Z\",\"show_cover_pic\":0}}",
			wantErr: true},
		{name: "No date", fields: fields{Path: "tmp"}, vector: "{\"status\":true,\"data\":{\"need_open_comment\":1,\"content\":\"<section></section>\",\"content_source_url\":\"https://m.joudou.com/invwinner#/winner?page=term&termid=1\",\"digest\":\"挥别四年一轮的二二九\",\"title\":\"【刊外四十】玩转双增长\",\"author\":\"骑行客\",\"only_fans_can_comment\":0,\"url\":\"http://mp.weixin.qq.com/s?__biz=Mzg3MTIzNTkyMw==&mid=100000703&idx=1&sn=5d92abf3f1c31eb52303c301d15bc016&chksm=4e80eb8379f76295c63cc75b48f890991c7238c6dcdf749dc060a09f70fbe8dbb4a9de0d894d#rd\",\"ext_info\":{\"read_count\":0,\"like_count\":0,\"name\":\"霄云笔谈\",\"prdname\":\"刊外\",\"jumpurl\":\"https://m.joudou.com/invwinner-kanwai/6422658b16dd4e0fb3cb2ad4ab7b648a\",\"mp_name\":\"骑行夜幕统计客\",\"mp_url\":\"https://mp.weixin.qq.com/mp/profile_ext?action=home&__biz=Mzg3MTIzNTkyMw==&scene=110#wechat_redirect\",\"icon\":\"https://static.joudou.com/wechat/dknight_avatar.png\"},\"thumb_media_id\":\"PNE0wI8i_2_xJurtZ7ztbj-wBkXRjT6H72nX-XM9gnw\",\"thumb_url\":\"http://mmbiz.qpic.cn/mmbiz_jpg/BOQXE8706a8mic3ibMiaNy0XfMam07N0JuBvj4jv2E3L2eFg8FibA78ehx4iayAXetECiaU0preQRPjevZoKXDaJTibOQ/0?wx_fmt=jpeg\",\"releasee\":\"2020-02-29T16:35:33Z\",\"show_cover_pic\":0}}",
			wantErr: true},
		{name: "Bad date", fields: fields{Path: "tmp"}, vector: "{\"status\":true,\"data\":{\"need_open_comment\":1,\"content\":\"<section></section>\",\"content_source_url\":\"https://m.joudou.com/invwinner#/winner?page=term&termid=1\",\"digest\":\"挥别四年一轮的二二九\",\"title\":\"【刊外四十】玩转双增长\",\"author\":\"骑行客\",\"only_fans_can_comment\":0,\"url\":\"http://mp.weixin.qq.com/s?__biz=Mzg3MTIzNTkyMw==&mid=100000703&idx=1&sn=5d92abf3f1c31eb52303c301d15bc016&chksm=4e80eb8379f76295c63cc75b48f890991c7238c6dcdf749dc060a09f70fbe8dbb4a9de0d894d#rd\",\"ext_info\":{\"read_count\":0,\"like_count\":0,\"name\":\"霄云笔谈\",\"prdname\":\"刊外\",\"jumpurl\":\"https://m.joudou.com/invwinner-kanwai/6422658b16dd4e0fb3cb2ad4ab7b648a\",\"mp_name\":\"骑行夜幕统计客\",\"mp_url\":\"https://mp.weixin.qq.com/mp/profile_ext?action=home&__biz=Mzg3MTIzNTkyMw==&scene=110#wechat_redirect\",\"icon\":\"https://static.joudou.com/wechat/dknight_avatar.png\"},\"thumb_media_id\":\"PNE0wI8i_2_xJurtZ7ztbj-wBkXRjT6H72nX-XM9gnw\",\"thumb_url\":\"http://mmbiz.qpic.cn/mmbiz_jpg/BOQXE8706a8mic3ibMiaNy0XfMam07N0JuBvj4jv2E3L2eFg8FibA78ehx4iayAXetECiaU0preQRPjevZoKXDaJTibOQ/0?wx_fmt=jpeg\",\"releasedate\":\"16:35:33Z\",\"show_cover_pic\":0}}",
			wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &kwWriter{
				Path: tt.fields.Path,
				buf:  tt.fields.buf,
			}
			k.Write([]byte(tt.vector))

			if err := k.Close(); (err != nil) != tt.wantErr {
				t.Errorf("kwWriter.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if err := os.RemoveAll("tmp"); err != nil {
		log.Fatal(err)
	}
}

func Test_genContent(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Good img", args: args{
			content: "<img data-src=\"https://mmbiz.qpic.cn/mmbiz_png/8rlZoIeurNvKNzrUiaKyvEn4vsCjqAnEiaHHGia8PkRBenP1CBddBQH5HkLcPiaqyZMlfeeAorGPKdAttCXicaV7mzw/640?wx_fmt=png\"  />",
		}, want: "<img src=\"https://mmbiz.qpic.cn/mmbiz_png/8rlZoIeurNvKNzrUiaKyvEn4vsCjqAnEiaHHGia8PkRBenP1CBddBQH5HkLcPiaqyZMlfeeAorGPKdAttCXicaV7mzw/640?wx_fmt=png\"  />"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genContent(tt.args.content); got != tt.want {
				t.Errorf("genContent() = %v, want %v", got, tt.want)
			}
		})
	}
}
