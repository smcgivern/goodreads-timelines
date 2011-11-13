set(:views, Proc.new {File.join(root, 'view')})
set(:haml, :format => :html5)

ROOT = (File.exist?('.root') ?
        File.open('.root').read.strip : nil)

helpers do
  # Call a partial HAML template.
  def partial(template, variables={})
    haml(template, {:layout => false}, variables)
  end

  # Prepend ROOT (from .root, if it exists) to absolute paths.
  def r(s)
    (s =~ /^\// && !(s =~ /^\/#{ROOT}/)) ? "/#{ROOT}#{s}" : s
  end

  # Take the inner_text and remove surrounding whitespace from a
  # Nokogiri element.
  def t(e); e.inner_text.strip; end
end

get('/ext/style.css') {scss(:style)}

get('/') do
  @page_title = 'Goodreads timelines'
  @scripts = ['/ext/jquery-1.7.min.js', '/ext/index.js']

  haml(:index)
end

post('/go-to-timeline/?') do
  user_id = Goodreads.find_user_id(params['goodreads-uri'])

  redirect(r("/:#{user_id}/"))
end

get('/::user_id/?') do
  @user_id = params['user_id']
  @user_info = Goodreads.user_info(@user_id)
  @all_reviews = Goodreads.all_reviews(@user_id)
  @by_day = Goodreads.reviews_by_date(@all_reviews)

  @by_month = Goodreads.reviews_by_date(@all_reviews, fill=:month)
  @by_month = Hash[@by_month.
                   group_by {|d, r| d.strftime("%Y/%m")}.
                   map {|k, v| [k, Hash[v]]}]

  haml(:timeline)
end
