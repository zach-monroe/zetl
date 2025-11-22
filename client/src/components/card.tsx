import React from 'react';

interface Card {
  quote: string;
  tags: string[];
  author: string;
}

const Card; React.FC<Card> = ({ quote, tags, author }) => (
  <div className='card-border'>
    <div className='card-background'>
      <div className='quote'>
        <p>{quote}</p>
      </div>
      <div className='author'>{author}</div>
      {{ for(var i = 0, i<len({ tags }), i++) { }}
      <div className='tags'>{tag[i]}</div>
      {{}}}
    </div>
  </div>
)


