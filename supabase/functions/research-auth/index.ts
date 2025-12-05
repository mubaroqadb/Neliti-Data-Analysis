// Research Data Analysis - Authentication Edge Function
// Handles user registration, login, and profile management

Deno.serve(async (req) => {
  const corsHeaders = {
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Headers': 'authorization, x-client-info, apikey, content-type',
    'Access-Control-Allow-Methods': 'POST, GET, OPTIONS',
    'Access-Control-Max-Age': '86400',
  };

  if (req.method === 'OPTIONS') {
    return new Response(null, { status: 200, headers: corsHeaders });
  }

  try {
    const url = new URL(req.url);
    const path = url.pathname.split('/').pop();
    
    const supabaseUrl = Deno.env.get('SUPABASE_URL');
    const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');
    
    if (!supabaseUrl || !serviceRoleKey) {
      throw new Error('Supabase configuration missing');
    }

    // Handle different auth endpoints
    if (req.method === 'POST') {
      const body = await req.json();
      
      if (path === 'register' || url.searchParams.get('action') === 'register') {
        // Registration
        const { email, password, full_name, institution, research_field } = body;
        
        if (!email || !password || !full_name) {
          return new Response(JSON.stringify({ 
            error: { code: 'VALIDATION_ERROR', message: 'Email, password, dan nama lengkap wajib diisi' }
          }), { status: 400, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
        }
        
        // Hash password using Web Crypto API
        const encoder = new TextEncoder();
        const data = encoder.encode(password + email);
        const hashBuffer = await crypto.subtle.digest('SHA-256', data);
        const hashArray = Array.from(new Uint8Array(hashBuffer));
        const passwordHash = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
        
        // Check if user exists
        const checkResponse = await fetch(`${supabaseUrl}/rest/v1/research_users?email=eq.${encodeURIComponent(email)}`, {
          headers: {
            'Authorization': `Bearer ${serviceRoleKey}`,
            'apikey': serviceRoleKey,
          }
        });
        
        const existingUsers = await checkResponse.json();
        if (existingUsers && existingUsers.length > 0) {
          return new Response(JSON.stringify({ 
            error: { code: 'USER_EXISTS', message: 'Email sudah terdaftar' }
          }), { status: 400, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
        }
        
        // Create user
        const createResponse = await fetch(`${supabaseUrl}/rest/v1/research_users`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${serviceRoleKey}`,
            'apikey': serviceRoleKey,
            'Content-Type': 'application/json',
            'Prefer': 'return=representation'
          },
          body: JSON.stringify({
            email,
            full_name,
            password_hash: passwordHash,
            institution: institution || null,
            research_field: research_field || null
          })
        });
        
        if (!createResponse.ok) {
          const errorText = await createResponse.text();
          throw new Error(`Registration failed: ${errorText}`);
        }
        
        const userData = await createResponse.json();
        const user = userData[0];
        
        // Generate simple token (in production, use proper JWT/PASETO)
        const tokenData = { user_id: user.id, email: user.email, exp: Date.now() + 86400000 };
        const token = btoa(JSON.stringify(tokenData));
        
        return new Response(JSON.stringify({
          data: {
            user: { id: user.id, email: user.email, full_name: user.full_name },
            token
          }
        }), { headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
        
      } else if (path === 'login' || url.searchParams.get('action') === 'login') {
        // Login
        const { email, password } = body;
        
        if (!email || !password) {
          return new Response(JSON.stringify({ 
            error: { code: 'VALIDATION_ERROR', message: 'Email dan password wajib diisi' }
          }), { status: 400, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
        }
        
        // Hash password
        const encoder = new TextEncoder();
        const data = encoder.encode(password + email);
        const hashBuffer = await crypto.subtle.digest('SHA-256', data);
        const hashArray = Array.from(new Uint8Array(hashBuffer));
        const passwordHash = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
        
        // Find user
        const findResponse = await fetch(
          `${supabaseUrl}/rest/v1/research_users?email=eq.${encodeURIComponent(email)}&password_hash=eq.${passwordHash}`,
          {
            headers: {
              'Authorization': `Bearer ${serviceRoleKey}`,
              'apikey': serviceRoleKey,
            }
          }
        );
        
        const users = await findResponse.json();
        if (!users || users.length === 0) {
          return new Response(JSON.stringify({ 
            error: { code: 'INVALID_CREDENTIALS', message: 'Email atau password salah' }
          }), { status: 401, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
        }
        
        const user = users[0];
        const tokenData = { user_id: user.id, email: user.email, exp: Date.now() + 86400000 };
        const token = btoa(JSON.stringify(tokenData));
        
        return new Response(JSON.stringify({
          data: {
            user: { id: user.id, email: user.email, full_name: user.full_name },
            token
          }
        }), { headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
      }
    }
    
    if (req.method === 'GET') {
      // Get profile
      const authHeader = req.headers.get('authorization');
      if (!authHeader) {
        return new Response(JSON.stringify({ 
          error: { code: 'UNAUTHORIZED', message: 'Token tidak ditemukan' }
        }), { status: 401, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
      }
      
      const token = authHeader.replace('Bearer ', '');
      let tokenData;
      try {
        tokenData = JSON.parse(atob(token));
      } catch {
        return new Response(JSON.stringify({ 
          error: { code: 'INVALID_TOKEN', message: 'Token tidak valid' }
        }), { status: 401, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
      }
      
      if (tokenData.exp < Date.now()) {
        return new Response(JSON.stringify({ 
          error: { code: 'TOKEN_EXPIRED', message: 'Token sudah kadaluarsa' }
        }), { status: 401, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
      }
      
      const profileResponse = await fetch(
        `${supabaseUrl}/rest/v1/research_users?id=eq.${tokenData.user_id}&select=id,email,full_name,institution,research_field,created_at`,
        {
          headers: {
            'Authorization': `Bearer ${serviceRoleKey}`,
            'apikey': serviceRoleKey,
          }
        }
      );
      
      const profiles = await profileResponse.json();
      if (!profiles || profiles.length === 0) {
        return new Response(JSON.stringify({ 
          error: { code: 'USER_NOT_FOUND', message: 'User tidak ditemukan' }
        }), { status: 404, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
      }
      
      return new Response(JSON.stringify({ data: profiles[0] }), {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
      });
    }

    return new Response(JSON.stringify({ 
      error: { code: 'METHOD_NOT_ALLOWED', message: 'Method tidak diizinkan' }
    }), { status: 405, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });

  } catch (error) {
    console.error('Auth error:', error);
    return new Response(JSON.stringify({
      error: { code: 'AUTH_ERROR', message: error.message }
    }), { status: 500, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
  }
});
