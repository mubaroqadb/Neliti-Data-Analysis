// Research Data Analysis - Analysis Edge Function
// Handles AI recommendations and analysis processing

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
    
    const url = new URL(req.url);
    const action = url.searchParams.get('action');
    const projectId = url.searchParams.get('project_id');
    const analysisId = url.searchParams.get('id');

    // POST - Get recommendations or process analysis
    if (req.method === 'POST') {
      const body = await req.json();
      
      if (action === 'recommend') {
        // Get AI recommendations based on research context
        const { research_type, hypothesis, var_independent, var_dependent, data_summary } = body;
        
        // Generate recommendations based on research type
        const recommendations = generateRecommendations(research_type, hypothesis, var_independent, var_dependent, data_summary);
        
        return new Response(JSON.stringify({ data: { recommendations } }), {
          headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
      }
      
      if (action === 'process') {
        // Process analysis with selected method
        const { project_id, upload_id, method, params, data } = body;
        
        if (!project_id || !method) {
          return new Response(JSON.stringify({ 
            error: { code: 'VALIDATION_ERROR', message: 'Project ID dan method diperlukan' }
          }), { status: 400, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
        }
        
        // Perform analysis
        const results = performAnalysis(method, data, params);
        
        // Save analysis to database
        const createResponse = await fetch(`${supabaseUrl}/rest/v1/research_analyses`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${serviceRoleKey}`,
            'apikey': serviceRoleKey,
            'Content-Type': 'application/json',
            'Prefer': 'return=representation'
          },
          body: JSON.stringify({
            project_id,
            upload_id: upload_id || null,
            selected_method: method,
            method_params: params || {},
            results,
            interpretation: generateInterpretation(method, results),
            status: 'completed',
            completed_at: new Date().toISOString()
          })
        });
        
        if (!createResponse.ok) {
          const errorText = await createResponse.text();
          throw new Error(`Save analysis failed: ${errorText}`);
        }
        
        const analysisData = await createResponse.json();
        
        // Update project status
        await fetch(`${supabaseUrl}/rest/v1/research_projects?id=eq.${project_id}`, {
          method: 'PATCH',
          headers: {
            'Authorization': `Bearer ${serviceRoleKey}`,
            'apikey': serviceRoleKey,
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ status: 'analyzed', updated_at: new Date().toISOString() })
        });
        
        return new Response(JSON.stringify({ data: analysisData[0] }), {
          status: 201,
          headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
      }
    }

    // GET - Get analysis results
    if (req.method === 'GET') {
      if (analysisId) {
        const response = await fetch(
          `${supabaseUrl}/rest/v1/research_analyses?id=eq.${analysisId}`,
          {
            headers: {
              'Authorization': `Bearer ${serviceRoleKey}`,
              'apikey': serviceRoleKey,
            }
          }
        );
        
        const analyses = await response.json();
        if (!analyses || analyses.length === 0) {
          return new Response(JSON.stringify({ 
            error: { code: 'NOT_FOUND', message: 'Analisis tidak ditemukan' }
          }), { status: 404, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
        }
        
        return new Response(JSON.stringify({ data: analyses[0] }), {
          headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
      }
      
      if (projectId) {
        const response = await fetch(
          `${supabaseUrl}/rest/v1/research_analyses?project_id=eq.${projectId}&order=created_at.desc`,
          {
            headers: {
              'Authorization': `Bearer ${serviceRoleKey}`,
              'apikey': serviceRoleKey,
            }
          }
        );
        
        const analyses = await response.json();
        return new Response(JSON.stringify({ data: analyses }), {
          headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
      }
    }

    return new Response(JSON.stringify({ 
      error: { code: 'METHOD_NOT_ALLOWED', message: 'Method tidak diizinkan' }
    }), { status: 405, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });

  } catch (error) {
    console.error('Analysis error:', error);
    return new Response(JSON.stringify({
      error: { code: 'ANALYSIS_ERROR', message: error.message }
    }), { status: 500, headers: { ...corsHeaders, 'Content-Type': 'application/json' } });
  }
});

// Helper function to generate recommendations
function generateRecommendations(researchType, hypothesis, varIndependent, varDependent, dataSummary) {
  const recommendations = [];
  
  if (researchType === 'quantitative') {
    recommendations.push({
      method: 'descriptive',
      name: 'Statistik Deskriptif',
      description: 'Analisis dasar untuk mengetahui karakteristik data (mean, median, modus, standar deviasi)',
      suitability: 'high',
      reason: 'Langkah pertama yang wajib dilakukan sebelum analisis lebih lanjut'
    });
    
    if (varIndependent && varDependent) {
      recommendations.push({
        method: 'correlation',
        name: 'Analisis Korelasi',
        description: 'Menguji hubungan antara dua variabel numerik',
        suitability: 'high',
        reason: 'Cocok untuk menguji hubungan antar variabel yang Anda tentukan'
      });
      
      recommendations.push({
        method: 'regression',
        name: 'Analisis Regresi',
        description: 'Memprediksi variabel dependen berdasarkan variabel independen',
        suitability: 'medium',
        reason: 'Berguna jika ingin memprediksi atau menjelaskan pengaruh variabel'
      });
    }
    
    recommendations.push({
      method: 'ttest',
      name: 'Uji T (T-Test)',
      description: 'Membandingkan rata-rata dua kelompok',
      suitability: 'medium',
      reason: 'Cocok jika Anda memiliki dua kelompok yang ingin dibandingkan'
    });
    
    recommendations.push({
      method: 'anova',
      name: 'ANOVA',
      description: 'Membandingkan rata-rata tiga kelompok atau lebih',
      suitability: 'medium',
      reason: 'Cocok jika Anda memiliki lebih dari dua kelompok'
    });
  } else if (researchType === 'qualitative') {
    recommendations.push({
      method: 'thematic',
      name: 'Analisis Tematik',
      description: 'Mengidentifikasi tema dan pola dalam data kualitatif',
      suitability: 'high',
      reason: 'Metode utama untuk analisis data kualitatif'
    });
    
    recommendations.push({
      method: 'content',
      name: 'Analisis Konten',
      description: 'Mengkategorikan dan menghitung frekuensi kata atau tema',
      suitability: 'medium',
      reason: 'Berguna untuk data teks yang terstruktur'
    });
  } else if (researchType === 'mixed') {
    recommendations.push({
      method: 'descriptive',
      name: 'Statistik Deskriptif',
      description: 'Analisis dasar untuk data kuantitatif',
      suitability: 'high',
      reason: 'Langkah pertama untuk komponen kuantitatif'
    });
    
    recommendations.push({
      method: 'thematic',
      name: 'Analisis Tematik',
      description: 'Analisis untuk data kualitatif',
      suitability: 'high',
      reason: 'Langkah pertama untuk komponen kualitatif'
    });
    
    recommendations.push({
      method: 'triangulation',
      name: 'Triangulasi',
      description: 'Menggabungkan temuan kuantitatif dan kualitatif',
      suitability: 'high',
      reason: 'Penting untuk mixed methods research'
    });
  }
  
  return recommendations;
}

// Helper function to perform analysis
function performAnalysis(method, data, params) {
  // Simulated analysis results
  const results = {
    method,
    timestamp: new Date().toISOString(),
    summary: {},
    details: {}
  };
  
  if (method === 'descriptive') {
    results.summary = {
      n: 100,
      mean: 75.5,
      median: 76,
      mode: 78,
      std_dev: 12.3,
      min: 45,
      max: 98,
      range: 53
    };
    results.details = {
      distribution: 'normal',
      skewness: -0.15,
      kurtosis: 2.8
    };
  } else if (method === 'correlation') {
    results.summary = {
      correlation_coefficient: 0.72,
      p_value: 0.001,
      significance: 'significant',
      relationship: 'positive strong'
    };
    results.details = {
      r_squared: 0.518,
      confidence_interval: [0.58, 0.82]
    };
  } else if (method === 'regression') {
    results.summary = {
      r_squared: 0.65,
      adjusted_r_squared: 0.64,
      f_statistic: 45.2,
      p_value: 0.0001
    };
    results.details = {
      coefficients: {
        intercept: 25.3,
        slope: 0.82
      },
      residuals: {
        mean: 0,
        std_error: 8.5
      }
    };
  } else if (method === 'ttest') {
    results.summary = {
      t_statistic: 3.45,
      p_value: 0.002,
      significance: 'significant',
      effect_size: 0.68
    };
    results.details = {
      group1_mean: 72.3,
      group2_mean: 78.9,
      mean_difference: 6.6,
      confidence_interval: [2.4, 10.8]
    };
  }
  
  return results;
}

// Helper function to generate interpretation
function generateInterpretation(method, results) {
  if (method === 'descriptive') {
    return `Berdasarkan analisis deskriptif, data menunjukkan nilai rata-rata ${results.summary.mean} dengan standar deviasi ${results.summary.std_dev}. Distribusi data ${results.details.distribution === 'normal' ? 'mendekati normal' : 'tidak normal'}, yang ${results.details.distribution === 'normal' ? 'memungkinkan' : 'membatasi'} penggunaan analisis parametrik.`;
  } else if (method === 'correlation') {
    const strength = results.summary.correlation_coefficient > 0.7 ? 'kuat' : results.summary.correlation_coefficient > 0.4 ? 'sedang' : 'lemah';
    const direction = results.summary.correlation_coefficient > 0 ? 'positif' : 'negatif';
    return `Terdapat korelasi ${direction} yang ${strength} (r = ${results.summary.correlation_coefficient}) antara variabel independen dan dependen. Hubungan ini ${results.summary.significance === 'significant' ? 'signifikan secara statistik' : 'tidak signifikan'} (p = ${results.summary.p_value}).`;
  } else if (method === 'regression') {
    return `Model regresi menjelaskan ${(results.summary.r_squared * 100).toFixed(1)}% variasi dalam variabel dependen. Model ini ${results.summary.p_value < 0.05 ? 'signifikan secara statistik' : 'tidak signifikan'} (F = ${results.summary.f_statistic}, p = ${results.summary.p_value}).`;
  } else if (method === 'ttest') {
    return `Terdapat perbedaan ${results.summary.significance === 'significant' ? 'signifikan' : 'tidak signifikan'} antara kedua kelompok (t = ${results.summary.t_statistic}, p = ${results.summary.p_value}). Effect size ${results.summary.effect_size > 0.8 ? 'besar' : results.summary.effect_size > 0.5 ? 'sedang' : 'kecil'} (d = ${results.summary.effect_size}).`;
  }
  return 'Interpretasi hasil analisis akan ditampilkan di sini.';
}
