// 文件上传到Supabase Storage的Edge Function
Deno.serve(async (req) => {
    const corsHeaders = {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Headers': 'authorization, x-client-info, apikey, content-type',
        'Access-Control-Allow-Methods': 'POST, GET, OPTIONS, PUT, DELETE, PATCH',
        'Access-Control-Max-Age': '86400',
        'Access-Control-Allow-Credentials': 'false'
    };

    if (req.method === 'OPTIONS') {
        return new Response(null, { status: 200, headers: corsHeaders });
    }

    try {
        // 获取Supabase服务密钥
        const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');
        const supabaseUrl = Deno.env.get('SUPABASE_URL');
        
        if (!serviceRoleKey || !supabaseUrl) {
            throw new Error('Missing Supabase configuration');
        }

        // 读取HTML文件内容
        const htmlContent = `<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/vite.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>NOFX - AI Auto Trading Dashboard</title>
    <script type="module" crossorigin src="https://eqzurdzoaxibothslnna.supabase.co/storage/v1/object/public/nofx-frontend/assets/index-TqyJ7tQt.js"></script>
    <link rel="stylesheet" crossorigin href="https://eqzurdzoaxibothslnna.supabase.co/storage/v1/object/public/nofx-frontend/assets/index-D6d1LzQD.css">
  </head>
  <body>
    <div id="root"></div>
  </body>
</html>`;

        // 上传HTML文件
        const uploadHtml = await fetch(`${supabaseUrl}/storage/v1/object/nofx-frontend/index.html`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'Content-Type': 'text/html'
            },
            body: htmlContent
        });

        // 读取CSS文件
        const cssResponse = await fetch('/workspace/nofx/web/dist/assets/index-D6d1LzQD.css');
        const cssContent = await cssResponse.text();

        // 上传CSS文件
        const uploadCss = await fetch(`${supabaseUrl}/storage/v1/object/nofx-frontend/assets/index-D6d1LzQD.css`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'Content-Type': 'text/css'
            },
            body: cssContent
        });

        // 读取JS文件
        const jsResponse = await fetch('/workspace/nofx/web/dist/assets/index-TqyJ7tQt.js');
        const jsContent = await jsResponse.text();

        // 上传JS文件
        const uploadJs = await fetch(`${supabaseUrl}/storage/v1/object/nofx-frontend/assets/index-TqyJ7tQt.js`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'Content-Type': 'application/javascript'
            },
            body: jsContent
        });

        const results = {
            html: uploadHtml.ok,
            css: uploadCss.ok,
            js: uploadJs.ok,
            publicUrl: `https://eqzurdzoaxibothslnna.supabase.co/storage/v1/object/public/nofx-frontend/index.html`
        };

        return new Response(JSON.stringify({
            success: true,
            message: 'Files uploaded successfully',
            data: results
        }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });

    } catch (error) {
        return new Response(JSON.stringify({
            success: false,
            error: error.message
        }), {
            status: 500,
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
    }
});