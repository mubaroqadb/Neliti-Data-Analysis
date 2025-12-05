// Research Data Analysis - Project Management Edge Function
// Handles CRUD operations for research projects

Deno.serve(async (req) => {
  const corsHeaders = {
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Headers': 'authorization, x-client-info, apikey, content-type',
    'Access-Control-Allow-Methods': 'POST, GET, PUT, DELETE, OPTIONS',
    'Access-Control-Max-Age': '86400',
  };

  if (req.method === 'OPTIONS') {
    return new Response(null, { status: 200, headers: corsHeaders });
  }

  try {
    const supabaseUrl = Deno.env.get('SUPABASE_URL');
    const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');
    
    if (!supabaseUrl || !serviceRoleKey) {
      throw new Error('Supabase configuration missing');
    }

    // Verify token
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
    
    const userId = tokenData.user_id;
    const url = new URL(req.url);
    const projectId = url.searchParams.get('id');

    // GET - List or get single project
    if (req.method === 'GET') {
      let query = `${supabaseUrl}/rest/v1/research_projects?user_id=eq.${userId}&order=created_at.desc`;
      if (projectId) {
        query = `${supabaseUrl}/rest/v1/research_projects?id=eq.${projectId}&user_id=eq.${userId}`;
      }
      
      const response = await fetch(query, {
        headers: {
          'Authorization': `Bearer ${serviceRoleKey}`,
          'apikey': serviceRoleKey,
        }
      });
      
      const projects = await response.json();
      
      if (projectId) {
        if (!projects || projects.length === 0) {
          return new Response(JSON.stringify({ 
            error: { code: 'NOT_FOUND', message: 'Proyek tidak ditemukan' }
          }), { status: 404, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
        }
        return new Response(JSON.stringify({ data: projects[0] }), {
          headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
      }
      
      return new Response(JSON.stringify({ data: projects }), {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
      });
    }

    // POST - Create project
    if (req.method === 'POST') {
      const body = await req.json();
      const { title, description, research_type, hypothesis, var_independent, var_dependent } = body;
      
      if (!title || !research_type) {
        return new Response(JSON.stringify({ 
          error: { code: 'VALIDATION_ERROR', message: 'Judul dan jenis penelitian wajib diisi' }
        }), { status: 400, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
      }
      
      const createResponse = await fetch(`${supabaseUrl}/rest/v1/research_projects`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${serviceRoleKey}`,
          'apikey': serviceRoleKey,
          'Content-Type': 'application/json',
          'Prefer': 'return=representation'
        },
        body: JSON.stringify({
          user_id: userId,
          title,
          description: description || null,
          research_type,
          hypothesis: hypothesis || null,
          var_independent: var_independent || null,
          var_dependent: var_dependent || null,
          status: 'draft'
        })
      });
      
      if (!createResponse.ok) {
        const errorText = await createResponse.text();
        throw new Error(`Create project failed: ${errorText}`);
      }
      
      const projectData = await createResponse.json();
      return new Response(JSON.stringify({ data: projectData[0] }), {
        status: 201,
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
      });
    }

    // PUT - Update project
    if (req.method === 'PUT') {
      if (!projectId) {
        return new Response(JSON.stringify({ 
          error: { code: 'VALIDATION_ERROR', message: 'Project ID diperlukan' }
        }), { status: 400, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
      }
      
      const body = await req.json();
      const updateData = { ...body, updated_at: new Date().toISOString() };
      delete updateData.id;
      delete updateData.user_id;
      delete updateData.created_at;
      
      const updateResponse = await fetch(
        `${supabaseUrl}/rest/v1/research_projects?id=eq.${projectId}&user_id=eq.${userId}`,
        {
          method: 'PATCH',
          headers: {
            'Authorization': `Bearer ${serviceRoleKey}`,
            'apikey': serviceRoleKey,
            'Content-Type': 'application/json',
            'Prefer': 'return=representation'
          },
          body: JSON.stringify(updateData)
        }
      );
      
      if (!updateResponse.ok) {
        const errorText = await updateResponse.text();
        throw new Error(`Update project failed: ${errorText}`);
      }
      
      const updatedData = await updateResponse.json();
      if (!updatedData || updatedData.length === 0) {
        return new Response(JSON.stringify({ 
          error: { code: 'NOT_FOUND', message: 'Proyek tidak ditemukan' }
        }), { status: 404, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
      }
      
      return new Response(JSON.stringify({ data: updatedData[0] }), {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
      });
    }

    // DELETE - Delete project
    if (req.method === 'DELETE') {
      if (!projectId) {
        return new Response(JSON.stringify({ 
          error: { code: 'VALIDATION_ERROR', message: 'Project ID diperlukan' }
        }), { status: 400, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
      }
      
      const deleteResponse = await fetch(
        `${supabaseUrl}/rest/v1/research_projects?id=eq.${projectId}&user_id=eq.${userId}`,
        {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${serviceRoleKey}`,
            'apikey': serviceRoleKey,
          }
        }
      );
      
      if (!deleteResponse.ok) {
        const errorText = await deleteResponse.text();
        throw new Error(`Delete project failed: ${errorText}`);
      }
      
      return new Response(JSON.stringify({ data: { success: true } }), {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
      });
    }

    return new Response(JSON.stringify({ 
      error: { code: 'METHOD_NOT_ALLOWED', message: 'Method tidak diizinkan' }
    }), { status: 405, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });

  } catch (error) {
    console.error('Project error:', error);
    return new Response(JSON.stringify({
      error: { code: 'PROJECT_ERROR', message: error.message }
    }), { status: 500, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
  }
});
